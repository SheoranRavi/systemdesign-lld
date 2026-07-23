package util

const RULES_KEY string = "/ratelimit/rules/"

func GetRuleKey(ruleBucket string, path string) string {
	return RULES_KEY + ruleBucket + path
}
