package utils

type AppKey struct {
	Appkey     string   `json:"appkey"     yaml:"appkey"`
	Secret     string   `json:"secret"     yaml:"secret"`
	Permission []string `json:"permission" yaml:"permission"`
}
