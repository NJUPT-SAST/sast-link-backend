package util

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"math/big"
	mr "math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/google/uuid"
)

const (
	larkSMTPServer = "smtp.feishu.cn:465"
)

func OutputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

// Generate UUID.
func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GenerateRandomString return a random string with length n.
func GenerateRandomString(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		// The reason for using crypto/rand instead of math/rand is that
		// the former relies on hardware to generate random numbers and
		// thus has a stronger source of random numbers.
		randNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		if _, err := sb.WriteRune(letters[randNum.Uint64()]); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

// GenerateCode generate a random code.
func GenerateCode() string {
	seed := time.Now().UnixNano() + int64(mr.Intn(4478))
	// rand.NewSource()
	rand := mr.New(mr.NewSource(seed))
	// 除去容易混淆的字符
	chars := "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

	// 生成6位随机数
	code := make([]byte, 5)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}

	// 将字符数组转换为字符串，并将第一个字符设置为"S"
	codeStr := fmt.Sprintf("S-%s", string(code))
	return codeStr
}

// sendEmail send email to user.
func SendEmail(sender, secret, recipient, content, title string) error {
	// https://gist.github.com/chrisgillis/10888032
	from := mail.Address{
		Name:    "SAST-Link",
		Address: sender,
	}
	to := mail.Address{
		Name:    "",
		Address: recipient,
	}
	body := content

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = title

	// setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	header := "Content-Type: text/html; charset=\"UTF-8\";\n\n"
	message += header + "\r\n" + body

	// Connect to the SMTP server
	servername := larkSMTPServer

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", sender, secret, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func UserNameToEmail(username string) string {
	return username + "@njupt.edu.cn"
}

// ShaHashing use sha512 to hash input.
func ShaHashing(in string) string {
	sha512Hash := sha512.Sum512([]byte(in))
	return hex.EncodeToString(sha512Hash[:])
}

// GetStudentIDFromEmail get student id from email
//
// The student id is the part before the '@' in the email. The student id is lowercase.
func GetStudentIDFromEmail(email string) string {
	if !strings.Contains(email, "@") {
		return ""
	}

	split := regexp.MustCompile(`@`)
	uid := split.Split(email, 2)[0]
	// Lowercase the uid
	return strings.ToLower(uid)
}

// ImageToWebp converts an image to WebP format and returns the result as a byte array.
//
// quality is a float32 value between 0 and 100. A quality of 75 is recommended.
func ImageToWebp(data []byte, quality float32) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: quality})
	if err != nil {
		return nil, fmt.Errorf("error encoding image to WebP: %w", err)
	}

	return buf.Bytes(), nil
}

// MapToJSONString converts a map to a JSON string.
func MapToJSONString(m map[string]string) (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func MapToJSONStringInterface(m map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// maskSecret masks the middle part of the secret, showing only the first and last 4 characters.
// If the secret is shorter than 8 characters, it replaces the whole secret with "*******".
func MaskSecret(secret string) string {
	if len(secret) > 8 {
		return secret[:4] + "*******" + secret[len(secret)-4:]
	}
	return "*******"
}

// CheckRedirectURI checks whether the redirect URI is valid.
func CheckRedirectURI(uri string) bool {
	parseURI, err := url.Parse(uri)
	if err != nil {
		return false
	}

	if parseURI.Scheme == "" || parseURI.Host == "" {
		return false
	}

	if parseURI.Scheme != "http" && parseURI.Scheme != "https" {
		return false
	}

	return true
}
