package discovery

import "fmt"

func AddEndOrRemoveFirstSlashIfNeeded(addr string) string {
	b := []byte(addr)
	if len(b) == 0 {
		return ""
	}
	lastByte := b[len(b)-1]
	if lastByte != '/' {
		addr = fmt.Sprintf("%s/", addr)
	}
	if b[0] == '/' {
		b = append(b[1:])
		addr = string(b)
	}
	return addr
}
