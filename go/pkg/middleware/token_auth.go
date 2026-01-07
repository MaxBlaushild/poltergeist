package middleware

import (
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/http"
	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/gin-gonic/gin"
)

const (
	bearer = "Bearer"
)

func WithAuthentication(authClient auth.Client, livenessClient liveness.LivenessClient, next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.Request.Header.Get("Authorization")
		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != bearer {
			ctx.JSON(401, http.ErrorResponse{
				Error: "invalid authorization header",
			})
			return
		}

		user, err := authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{
			Token: headerParts[1],
		})
		if err != nil {
			ctx.JSON(401, http.ErrorResponse{
				Error: "authorization header not valid",
			})
			return
		}

		// Extract and save user location if provided
		locationHeader := ctx.Request.Header.Get("X-User-Location")
		log.Printf("[DEBUG] Location header: %s", locationHeader)
		if locationHeader != "" {
			if err = livenessClient.SetUserLocation(ctx, user.ID, locationHeader); err != nil {
				log.Println("error setting user location", err)
			}
		}

		ctx.Set("user", user)

		next(ctx)
	}
}

func WithAuthenticationWithoutLocation(authClient auth.Client, next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.Request.Header.Get("Authorization")
		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != bearer {
			ctx.JSON(401, http.ErrorResponse{
				Error: "invalid authorization header",
			})
			return
		}

		user, err := authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{
			Token: headerParts[1],
		})
		if err != nil {
			ctx.JSON(401, http.ErrorResponse{
				Error: "authorization header not valid",
			})
			return
		}

		ctx.Set("user", user)

		next(ctx)
	}
}

func WithAntiviralCookie(next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		antiviralParam := ctx.Query("antiviral")
		if antiviralParam != "true" {
			html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Antivirus Software Required</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Courier New', monospace;
            background: #000000;
            color: #00ff00;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            overflow: hidden;
            position: relative;
        }
        body::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: 
                repeating-linear-gradient(
                    0deg,
                    rgba(0, 255, 0, 0.03) 0px,
                    transparent 1px,
                    transparent 2px,
                    rgba(0, 255, 0, 0.03) 3px
                );
            pointer-events: none;
            animation: scan 8s linear infinite;
        }
        @keyframes scan {
            0% { transform: translateY(0); }
            100% { transform: translateY(20px); }
        }
        .container {
            background: rgba(0, 0, 0, 0.9);
            border: 2px solid #00ff00;
            border-radius: 8px;
            padding: 40px;
            max-width: 600px;
            width: 100%;
            box-shadow: 
                0 0 20px rgba(0, 255, 0, 0.5),
                inset 0 0 20px rgba(0, 255, 0, 0.1);
            position: relative;
            z-index: 1;
            animation: pulse 2s ease-in-out infinite;
        }
        @keyframes pulse {
            0%, 100% { box-shadow: 0 0 20px rgba(0, 255, 0, 0.5), inset 0 0 20px rgba(0, 255, 0, 0.1); }
            50% { box-shadow: 0 0 30px rgba(0, 255, 0, 0.8), inset 0 0 30px rgba(0, 255, 0, 0.2); }
        }
        h1 {
            font-size: 28px;
            margin-bottom: 20px;
            text-align: center;
            text-transform: uppercase;
            letter-spacing: 3px;
            text-shadow: 0 0 10px rgba(0, 255, 0, 0.8);
            animation: glitch 3s infinite;
        }
        @keyframes glitch {
            0%, 100% { transform: translate(0); }
            20% { transform: translate(-2px, 2px); }
            40% { transform: translate(-2px, -2px); }
            60% { transform: translate(2px, 2px); }
            80% { transform: translate(2px, -2px); }
        }
        .error-code {
            font-size: 14px;
            color: #ff0080;
            margin-bottom: 30px;
            text-align: center;
            font-weight: bold;
        }
        .message {
            font-size: 18px;
            line-height: 1.6;
            margin-bottom: 30px;
            text-align: center;
        }
        .highlight {
            color: #00ffff;
            font-weight: bold;
            text-shadow: 0 0 5px rgba(0, 255, 255, 0.8);
        }
        .warning {
            margin-top: 20px;
            padding: 15px;
            border-left: 4px solid #ff0080;
            background: rgba(255, 0, 128, 0.1);
            font-size: 14px;
            color: #ff0080;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚠ SYSTEM SECURITY ALERT ⚠</h1>
        <div class="error-code">ERROR: ANTIVIRUS SOFTWARE NOT DETECTED</div>
        <div class="message">
            You need to interface with this server using an <span class="highlight">antiviral software scanner program</span>.
        </div>
        <div class="warning">
            <strong>WARNING:</strong> Direct server interface without proper security protocols may result in system compromise.
        </div>
    </div>
</body>
</html>`
			ctx.Header("Content-Type", "text/html; charset=utf-8")
			ctx.String(403, html)
			return
		}

		next(ctx)
	}
}
