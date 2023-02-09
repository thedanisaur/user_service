package types

type Config struct {
	Database struct {
		ConfigPath string `json:"config_path"`
		DbConfig   struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Name     string `json:"db_name"`
			Driver   string `json:"db_driver"`
		}
	}
	App struct {
		Host struct {
			CertificatePath string `json:"cert_path"`
			KeyPath         string `json:"key_path"`
			Port            int    `json:"port"`
		}
		Cors struct {
			AllowCredentials bool     `json:"allow_credentials"`
			AllowHeaders     []string `json:"allow_headers"`
			AllowOrigins     []string `json:"allow_origins"`
		}
		Limiter struct {
			Max                      int  `json:"max_requests"`
			Expiration               int  `json:"expiration"`
			LimiterSlidingMiddleware bool `json:"limiter_sliding_middleware"`
			SkipSuccessfulRequests   bool `json:"skip_successful_requests"`
		}
	}
}
