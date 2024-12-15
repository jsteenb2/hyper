package main

import (
	"encoding/base64"
	"net/http"
	"time"
)

func setFlash(w http.ResponseWriter, name, value string) {
	c := &http.Cookie{Name: name, Value: encode(value), SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, c)
}

func getFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return "", nil
		default:
			return "", err
		}
	}
	value, err := decode(c.Value)
	if err != nil {
		return "", err
	}
	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0), SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, dc)
	return string(value), nil
}

func encode(src string) string {
	return base64.URLEncoding.EncodeToString([]byte(src))
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}
