package auth

func GenerateInviteLink(joinToken string) string {
	baseURL := "https://bb-in-view.web.app/?join="
	link := baseURL + joinToken
	return link
}
