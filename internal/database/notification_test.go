package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func queueNotification(t *testing.T, repo *database.UserRepo) database.Notification {
	t.Helper()
	user := database.NewUser("outbox-"+uuid.New().String()[:20], "Publisher", []byte("hash"), model.PublisherRoleName)
	user.SetEmail(uuid.New().String()+"@example.com", false)
	require.NoError(t, repo.CreateUser(t.Context(), user))
	v := database.NewEmailVerification(user.ID, "hash", "unsubscribe", time.Now().UTC())
	payload, err := json.Marshal(notification.Email{To: []string{user.Email.String}})
	require.NoError(t, err)
	n := database.Notification{ID: uuid.New().String(), ReferenceID: v.ID, Kind: "email_verification", Payload: payload,
		CreatedAt: v.DateCreated, ExpiresAt: v.DateCreated.Add(model.VerificationCodeTTL)}
	require.NoError(t, repo.RunWithTx(t.Context(), func(ctx context.Context) error {
		if err := repo.CreateEmailVerification(ctx, v); err != nil {
			return err
		}
		return repo.CreateNotification(ctx, n)
	}))
	return n
}

func TestNotificationClaimsAreExclusiveAndFenced(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	n := queueNotification(t, repo)
	var wg sync.WaitGroup
	claims := make(chan database.Notification, 2)
	errs := make(chan error, 2)
	for _, owner := range []string{"worker-a", "worker-b"} {
		wg.Go(func() {
			claim, err := repo.ClaimNotification(t.Context(), owner)
			if err == nil {
				claims <- claim
			} else {
				errs <- err
			}
		})
	}
	wg.Wait()
	require.Len(t, claims, 1)
	require.Len(t, errs, 1)
	require.ErrorIs(t, <-errs, database.ErrNotFound)
	old := <-claims
	require.Equal(t, n.ID, old.ID)
	_, err := db.Exec(`UPDATE notification_outbox SET locked_at = now() - interval '61 seconds' WHERE id = $1`, n.ID)
	require.NoError(t, err)
	current, err := repo.ClaimNotification(t.Context(), "worker-c")
	require.NoError(t, err)
	require.True(t, current.Reclaimed)
	require.ErrorIs(t, repo.CompleteNotification(t.Context(), old, "old-message"), database.ErrLeaseLost)
	require.ErrorIs(t, repo.RetryNotification(t.Context(), old, time.Second, "network"), database.ErrLeaseLost)
	require.NoError(t, repo.CompleteNotification(t.Context(), current, "message-id"))
	var got struct {
		Status    string `db:"status"`
		Payload   []byte `db:"payload"`
		MessageID string `db:"provider_message_id"`
	}
	require.NoError(t, db.Get(&got, `SELECT status, payload, provider_message_id FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, "sent", got.Status)
	require.Empty(t, got.Payload)
	require.Equal(t, "message-id", got.MessageID)
	var messageID string
	require.NoError(t, db.Get(&messageID, `SELECT message_id FROM email_verifications WHERE id=$1`, n.ReferenceID))
	require.Equal(t, "message-id", messageID)
}

func TestNotificationEnqueueRollbackAndExpiry(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	n := queueNotification(t, repo)
	wantErr := errors.New("rollback")
	err := repo.RunWithTx(t.Context(), func(ctx context.Context) error {
		if err := repo.SetEmailVerificationUsed(ctx, n.ReferenceID, false); err != nil {
			return err
		}
		duplicate := n
		duplicate.ID = uuid.New().String()
		// The unique verification reference rejects a second notification.
		if err := repo.CreateNotification(ctx, duplicate); err != nil {
			return wantErr
		}
		return nil
	})
	require.ErrorIs(t, err, wantErr)
	var code string
	require.NoError(t, db.Get(&code, `SELECT verification_code FROM email_verifications WHERE id=$1`, n.ReferenceID))
	require.Equal(t, "hash", code)
	_, err = db.Exec(`UPDATE notification_outbox SET expires_at=now()-interval '1 second', next_attempt_at=now()+interval '1 day' WHERE id=$1`, n.ID)
	require.NoError(t, err)
	require.NoError(t, repo.ExpireNotifications(t.Context()))
	var status string
	require.NoError(t, db.Get(&status, `SELECT status FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, "expired", status)
	_, err = repo.ClaimNotification(t.Context(), "worker")
	require.ErrorIs(t, err, database.ErrNotFound)
}

func TestConcurrentResendsQueueOnlyOneNotification(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	u := database.NewUser("publisher", "Publisher", []byte("hash"), model.PublisherRoleName)
	u.SetEmail("publisher@example.com", false)
	require.NoError(t, repo.CreateUser(t.Context(), u))
	renderer, err := resendapi.NewRenderer(resendapi.Config{FromEmail: "from@example.com", BaseURL: "https://example.com"})
	require.NoError(t, err)
	p := facade.New(zap.NewNop(), repo, renderer, nil, auth.NewUnsubscribeTokenGenerator([]byte("test")), nil)
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() { results <- p.ResendVerificationEmail(t.Context(), u.ID) })
	}
	wg.Wait()
	one, two := <-results, <-results
	if one != nil {
		one, two = two, one
	}
	require.NoError(t, one)
	require.NotNil(t, facade.AsTooManyRequestsError(two), "second resend returned: %v", two)
	var count int
	require.NoError(t, db.Get(&count, `SELECT count(*) FROM notification_outbox`))
	require.Equal(t, 1, count)
	var matches bool
	require.NoError(t, db.Get(&matches, `SELECT n.created_at=v.date_created AND n.expires_at=v.date_created+interval '24 hours'
		FROM notification_outbox n JOIN email_verifications v ON v.id::text=n.reference_id`))
	require.True(t, matches)
	n, err := repo.ClaimNotification(t.Context(), "worker")
	require.NoError(t, err)
	var email notification.Email
	require.NoError(t, json.Unmarshal(n.Payload, &email))
	require.Equal(t, []string{u.Email.String}, email.To)
}

func TestNotificationEnqueueFailureRollsBackNewVerification(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	existing := queueNotification(t, repo)
	var userID string
	require.NoError(t, db.Get(&userID, `SELECT user_id FROM email_verifications WHERE id=$1`, existing.ReferenceID))
	v := database.NewEmailVerification(userID, "new hash", "token", time.Now())
	err := repo.RunWithTx(t.Context(), func(ctx context.Context) error {
		if err := repo.CreateEmailVerification(ctx, v); err != nil {
			return err
		}
		n := existing
		n.ReferenceID = v.ID
		return repo.CreateNotification(ctx, n) // Duplicate primary key forces enqueue failure.
	})
	require.Error(t, err)
	var count int
	require.NoError(t, db.Get(&count, `SELECT count(*) FROM email_verifications WHERE id=$1`, v.ID))
	require.Zero(t, count)
}

func TestNotificationAcknowledgmentRollsBackBothWrites(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	queueNotification(t, repo)
	n, err := repo.ClaimNotification(t.Context(), "worker")
	require.NoError(t, err)
	_, err = db.Exec(`CREATE FUNCTION reject_notification_message_id() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'simulated verification write failure'; END $$;
		CREATE TRIGGER reject_notification_message_id BEFORE UPDATE OF message_id ON email_verifications
		FOR EACH ROW EXECUTE FUNCTION reject_notification_message_id()`)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec(`DROP TRIGGER reject_notification_message_id ON email_verifications; DROP FUNCTION reject_notification_message_id()`)
	}()
	require.Error(t, repo.CompleteNotification(t.Context(), n, "provider-id"))
	var status string
	var payload []byte
	require.NoError(t, db.Get(&status, `SELECT status FROM notification_outbox WHERE id=$1`, n.ID))
	require.NoError(t, db.Get(&payload, `SELECT payload FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, "pending", status)
	require.NotEmpty(t, payload)
}

func TestNotificationEligibilityAndActiveLeaseExpiry(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	queueNotification(t, repo)
	n, err := repo.ClaimNotification(t.Context(), "worker")
	require.NoError(t, err)
	var email string
	require.NoError(t, db.Get(&email, `SELECT u.email FROM users u JOIN email_verifications v ON v.user_id=u.id WHERE v.id=$1`, n.ReferenceID))
	eligible, err := repo.NotificationEligible(t.Context(), n, email)
	require.NoError(t, err)
	require.True(t, eligible)
	eligible, err = repo.NotificationEligible(t.Context(), n, "other@example.com")
	require.NoError(t, err)
	require.False(t, eligible, "recipient must match the user's email")
	require.NoError(t, repo.CreateEmailUnsubscribe(t.Context(), database.NewEmailUnsubscribe(email)))
	eligible, err = repo.NotificationEligible(t.Context(), n, email)
	require.NoError(t, err)
	require.False(t, eligible)
	_, err = db.Exec(`UPDATE notification_outbox SET expires_at=now()-interval '1 second' WHERE id=$1`, n.ID)
	require.NoError(t, err)
	require.NoError(t, repo.ExpireNotifications(t.Context()))
	var status string
	require.NoError(t, db.Get(&status, `SELECT status FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, "pending", status, "an active worker still owns this lease")
	_, err = db.Exec(`UPDATE notification_outbox SET locked_at=now()-interval '61 seconds' WHERE id=$1`, n.ID)
	require.NoError(t, err)
	require.NoError(t, repo.ExpireNotifications(t.Context()))
	require.NoError(t, db.Get(&status, `SELECT status FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, "expired", status)
}

func TestNotificationReferenceAndUpdateTime(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	now := time.Now().UTC().Add(-time.Minute)
	n := database.Notification{ID: uuid.New().String(), ReferenceID: "order-123", Kind: "order_receipt",
		Payload: []byte("{}"), CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	require.NoError(t, repo.CreateNotification(t.Context(), n), "no referenced entity is required")
	var created, updated time.Time
	require.NoError(t, db.Get(&created, `SELECT created_at FROM notification_outbox WHERE id=$1`, n.ID))
	require.NoError(t, db.Get(&updated, `SELECT updated_at FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, created, updated)
	claim, err := repo.ClaimNotification(t.Context(), "worker")
	require.NoError(t, err)
	eligible, err := repo.NotificationEligible(t.Context(), claim, "user@example.com")
	require.NoError(t, err)
	require.False(t, eligible, "other notification kinds do not use verification eligibility")
	require.NoError(t, db.Get(&updated, `SELECT updated_at FROM notification_outbox WHERE id=$1`, n.ID))
	require.True(t, updated.After(created))
	claimedAt := updated
	require.NoError(t, repo.RetryNotification(t.Context(), claim, time.Minute, "transport"))
	require.NoError(t, db.Get(&updated, `SELECT updated_at FROM notification_outbox WHERE id=$1`, n.ID))
	require.True(t, updated.After(claimedAt))
	_, err = repo.ClaimNotification(t.Context(), "worker")
	require.ErrorIs(t, err, database.ErrNotFound, "released lease must still respect next_attempt_at")
	_, err = db.Exec(`UPDATE notification_outbox SET next_attempt_at=now()-interval '1 second' WHERE id=$1`, n.ID)
	require.NoError(t, err)
	claim, err = repo.ClaimNotification(t.Context(), "worker")
	require.NoError(t, err)
	require.NoError(t, repo.CompleteNotification(t.Context(), claim, "provider-id"))
	var providerID string
	require.NoError(t, db.Get(&providerID, `SELECT provider_message_id FROM notification_outbox WHERE id=$1`, n.ID))
	require.Equal(t, "provider-id", providerID)
	duplicate := n
	duplicate.ID = uuid.New().String()
	require.Error(t, repo.CreateNotification(t.Context(), duplicate), "one notification per kind and reference")
	duplicate.Kind = "shipment_notice"
	require.NoError(t, repo.CreateNotification(t.Context(), duplicate), "different kinds may use the same reference")
	_, err = db.Exec(`UPDATE notification_outbox SET updated_at=now()-interval '31 days'`)
	require.NoError(t, err)
	require.NoError(t, repo.PurgeNotifications(t.Context()))
	var ids []string
	require.NoError(t, db.Select(&ids, `SELECT id FROM notification_outbox`))
	require.Equal(t, []string{duplicate.ID}, ids, "retention removes only terminal rows")
}

func TestNotificationMissingReferenceIsCancelled(t *testing.T) {
	repo := setup(t)
	defer teardown(t)
	n := queueNotification(t, repo)
	_, err := db.Exec(`DELETE FROM email_verifications WHERE id=$1`, n.ReferenceID)
	require.NoError(t, err)
	claim, err := repo.ClaimNotification(t.Context(), "worker")
	require.NoError(t, err, "deleting the source no longer cascades into the outbox")
	eligible, err := repo.NotificationEligible(t.Context(), claim, "user@example.com")
	require.NoError(t, err)
	require.False(t, eligible)
	require.NoError(t, repo.FinishNotification(t.Context(), claim, "cancelled", "obsolete_or_unsubscribed"))
	var payload []byte
	require.NoError(t, db.Get(&payload, `SELECT payload FROM notification_outbox WHERE id=$1`, n.ID))
	require.Empty(t, payload)
}
