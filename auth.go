package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

//用于在 hanlder 执行前检测 authcookie，类似于装饰器
type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")
	if err == http.ErrNoCookie || cookie.Value == "" {
		w.Header().Set("Location", "/login")
		//Writeheader用于写入一个 http status code，由于默认的是200，所以这个方法常用于返回错误 code
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.next.ServeHTTP(w, r)
}

// MustAuth 将 handler 转化为 authHandler 表示这个 handler 必须验证
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandler 如果不需要在调用时存储状态，可以直接用函数来 handle
func loginHanlder(w http.ResponseWriter, r *http.Request) {
	//由于 net/http 框架不支持类似	auth/{action}/{provider_name}的解析方式，因此这里手动拆解路径
	segs := strings.Split(r.URL.Path, "/")
	if len(segs) < 4 {
		http.Error(w, "Auth path wrong!", http.StatusBadRequest)
		return
	}
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while trying to get provider %s: %s", provider.Name(), err), http.StatusBadRequest)
			return
		}
		loginURL, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while trying to GetBeginAuthURL for %s: %s", provider.Name(), err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", loginURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while trying to get provider %s: %s", provider.Name(), err), http.StatusBadRequest)
			return
		}
		//在CompleteAuth里，可能遇到翻墙问题， Mac可以使用ShadowSocks + Proxifier 的方法解决
		log.Println("Completing Auth")
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		log.Println("Completed Auth")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while trying to CompleteAuth for %s: %s", provider.Name(), err), http.StatusInternalServerError)
			return
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while trying to GetUser for %s: %s", provider.Name(), err), http.StatusInternalServerError)
			return
		}
		//将用户名存储在一个 msi(map[string]interface()) 对象中，可以看做一个 JSON object。同时进行 base64 编码，方便传入 URL 或者存放在 cookie 中
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Email()))
		userID := fmt.Sprintf("%x", m.Sum(nil))
		authCookieValue := objx.New(map[string]interface{}{
			"userid":     userID,
			"name":       user.Name(),
			"avatar_url": user.AvatarURL(),
			"email":      user.Email(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supperted yet", action)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.Header().Set("Location", "/chat")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
