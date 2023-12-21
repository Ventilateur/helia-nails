package treatwell

type CookieStorer interface {
	SetCookie()
	RetrieveCookie()
}
