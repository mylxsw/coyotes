package middleware

import "net/http"

// WithJSONResponse 负责返回json响应
func WithJSONResponse(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		h(w, r)
	}
}

// WithHTMLResponse 负责返回html响应
func WithHTMLResponse(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		h(w, r)
	}
}

// HandlerDecorator 该函数是http handler的装饰器
type HandlerDecorator func(http.HandlerFunc) http.HandlerFunc

// Handler 用于包装http handler，对其进行装饰
func Handler(h http.HandlerFunc, decors ...HandlerDecorator) http.HandlerFunc {
	for i := range decors {
		d := decors[len(decors)-1-i]
		h = d(h)
	}

	return h
}
