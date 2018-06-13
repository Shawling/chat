package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/shawling/trace"
	"github.com/stretchr/objx"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ1   *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//模板只需要渲染一次。如果能确保只在一个 goroutine 中使用这个结构体，可以使用一些初始化代码。如果多个 goroutine ，可以使用 sync.Once
	//在调用时才渲染模板可以确保需要使用才渲染
	t.once.Do(func() {
		t.templ1 = template.Must(template.ParseFiles(filepath.Join("template", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	//由于用户信息来源于浏览器，因此只能在请求时才能获取用户信息并渲染
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)

		//获取用户头像
		client := new(client)
		client.userData = map[string]interface{}(data["UserData"].(objx.Map))
		avatarURL, _ := UseTryAvatars.GetAvatarURL(client)
		data["avatar"] = avatarURL
	}

	t.templ1.Execute(w, data)
}

func main() {
	//从命令行参数中获取 addr
	var addr = flag.String("addr", ":8080", "The addr of the  application.")
	flag.Parse()

	//setup gomniauth
	gomniauth.SetSecurityKey("98dfbg7iu2nb4uywevihjw4tuiyub34noilk")
	gomniauth.WithProviders(
		github.New("3d1e6ba69036e0624b61", "7e8938928d802e7582908a5eadaaaf22d64babf1", "http://localhost:8080/auth/callback/github"),
		google.New("44166123467-o6brs9o43tgaek9q12lef07bk48m3jmf.apps.googleusercontent.com", "rpXpakthfjPVoFGvcf9CVCu7", "http://localhost:8080/auth/callback/google"),
		facebook.New("537611606322077", "f9f4d77b3d3f4f5775369f5c9f88f65e", "http://localhost:8080/auth/callback/facebook"),
	)

	//route code
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))

	http.Handle("/login", &templateHandler{filename: "login.html"})

	http.Handle("/upload", MustAuth(&templateHandler{filename: "upload.html"}))

	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./avatars"))))

	http.HandleFunc("/auth/", loginHanlder)

	http.HandleFunc("/logout", logoutHandler)

	http.HandleFunc("/uploader", uploaderHandler)

	//creat new room bingding on a websocket address
	r := newRoom(UseTryAvatars)
	r.tracer = trace.New(os.Stdout)
	http.Handle("/room", r)
	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
