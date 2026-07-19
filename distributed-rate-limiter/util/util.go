package util

func GetRuleKey(ruleBucket string, path string) string {
	return "/ratelimit/rules/" + ruleBucket + "/" + path
}
