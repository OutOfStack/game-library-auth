// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/account": {
            "delete": {
                "description": "Deletes a user account",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Delete user account",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Successfully deleted account"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    }
                }
            },
            "patch": {
                "description": "Updates the profile information of a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Update user profile",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Update profile parameters",
                        "name": "profile",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.UpdateProfileReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Returns new access token",
                        "schema": {
                            "$ref": "#/definitions/handlers.TokenResp"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "401": {
                        "description": "Invalid password or token",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    }
                }
            }
        },
        "/oauth/google": {
            "post": {
                "description": "Handles Google OAuth 2.0 authentication",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Google OAuth sign in handler",
                "parameters": [
                    {
                        "description": "Google OAuth token",
                        "name": "token",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.GoogleOAuthRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.TokenResp"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    }
                }
            }
        },
        "/signin": {
            "post": {
                "description": "Authenticate a user and return an access token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Sign in",
                "parameters": [
                    {
                        "description": "User credentials",
                        "name": "signin",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.SignInReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.TokenResp"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    }
                }
            }
        },
        "/signup": {
            "post": {
                "description": "Create a new user account with the provided information",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Register a new user",
                "parameters": [
                    {
                        "description": "User signup information",
                        "name": "signup",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.SignUpReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully registered user",
                        "schema": {
                            "$ref": "#/definitions/handlers.SignUpResp"
                        }
                    },
                    "400": {
                        "description": "Invalid input data",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "409": {
                        "description": "Username or publisher name already exists",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    }
                }
            }
        },
        "/token/verify": {
            "post": {
                "description": "Validates a JWT token and returns if it's valid",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Verify JWT token",
                "parameters": [
                    {
                        "description": "Token to verify",
                        "name": "token",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.VerifyTokenReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.VerifyTokenResp"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.ErrResp"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.GoogleOAuthRequest": {
            "type": "object",
            "required": [
                "idToken"
            ],
            "properties": {
                "idToken": {
                    "type": "string"
                }
            }
        },
        "handlers.SignInReq": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 8
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "handlers.SignUpReq": {
            "type": "object",
            "required": [
                "name",
                "password",
                "username"
            ],
            "properties": {
                "confirmPassword": {
                    "type": "string"
                },
                "isPublisher": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 8
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "handlers.SignUpResp": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                }
            }
        },
        "handlers.TokenResp": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                }
            }
        },
        "handlers.UpdateProfileReq": {
            "type": "object",
            "properties": {
                "confirmNewPassword": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 8
                },
                "name": {
                    "type": "string"
                },
                "newPassword": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 8
                },
                "password": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 8
                }
            }
        },
        "handlers.VerifyTokenReq": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "handlers.VerifyTokenResp": {
            "type": "object",
            "properties": {
                "valid": {
                    "type": "boolean"
                }
            }
        },
        "web.ErrResp": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "fields": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.FieldError"
                    }
                }
            }
        },
        "web.FieldError": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "field": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.2",
	Host:             "localhost:8001",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Game library auth API",
	Description:      "API for game library auth service",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
