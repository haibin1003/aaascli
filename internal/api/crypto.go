package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// CryptoContext 保存加密通信所需的密钥对和平台公钥
type CryptoContext struct {
	LocalKeyPair   *RSAKeyPair
	PlatformPubKey string
	PubKeyExpireAt time.Time
}

// RSAKeyPair RSA 密钥对
type RSAKeyPair struct {
	PublicKey  string
	PrivateKey string
}

// GenerateRSAKeyPair 生成 2048 位 RSA 密钥对
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key failed: %w", err)
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public key failed: %w", err)
	}
	pubKeyPem := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyBytes}))

	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPem := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKeyBytes}))

	return &RSAKeyPair{
		PublicKey:  pubKeyPem,
		PrivateKey: privKeyPem,
	}, nil
}

// RSAEncrypt 用平台公钥加密数据（PKCS1v15 + Base64）
func RSAEncrypt(pubKeyPem, data string) (string, error) {
	block, _ := pem.Decode([]byte(pubKeyPem))
	if block == nil {
		return "", fmt.Errorf("failed to parse public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("parse public key failed: %w", err)
		}
	}
	rsaPub := pub.(*rsa.PublicKey)

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(data))
	if err != nil {
		return "", fmt.Errorf("RSA encrypt failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// RSADecrypt 用本地私钥解密数据（PKCS1v15 + Base64）
func RSADecrypt(privKeyPem, b64 string) (string, error) {
	block, _ := pem.Decode([]byte(privKeyPem))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key PEM")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse private key failed: %w", err)
	}

	encrypted, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encrypted)
	if err != nil {
		return "", fmt.Errorf("RSA decrypt failed: %w", err)
	}
	return string(decrypted), nil
}

// AESEncrypt AES-128-CBC-PKCS7 加密
// 返回格式: hex(iv) + base64(ciphertext)
func AESEncrypt(plainText string, key []byte) (string, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plainBytes := []byte(plainText)
	padLen := aes.BlockSize - len(plainBytes)%aes.BlockSize
	padded := append(plainBytes, make([]byte, padLen)...)
	for i := 0; i < padLen; i++ {
		padded[len(plainBytes)+i] = byte(padLen)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	return hex.EncodeToString(iv) + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES-128-CBC-PKCS7 解密
// 输入格式: hex(iv) + base64(ciphertext)
func AESDecrypt(cipherText string, key []byte) (string, error) {
	if len(cipherText) < 32 {
		return "", fmt.Errorf("cipherText too short")
	}
	iv, err := hex.DecodeString(cipherText[:32])
	if err != nil {
		return "", fmt.Errorf("decode iv failed: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(cipherText[32:])
	if err != nil {
		return "", fmt.Errorf("decode ciphertext failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher failed: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plainBytes := make([]byte, len(ciphertext))
	mode.CryptBlocks(plainBytes, ciphertext)

	padLen := int(plainBytes[len(plainBytes)-1])
	if padLen > aes.BlockSize || padLen == 0 {
		return "", fmt.Errorf("invalid padding")
	}
	plainBytes = plainBytes[:len(plainBytes)-padLen]

	return string(plainBytes), nil
}

// EncryptRequest 加密请求体
func (c *CryptoContext) EncryptRequest(body string, contentType string) (string, error) {
	if c.LocalKeyPair == nil {
		pair, err := GenerateRSAKeyPair()
		if err != nil {
			return "", err
		}
		c.LocalKeyPair = pair
	}

	if c.PlatformPubKey == "" || time.Now().After(c.PubKeyExpireAt) {
		return "", fmt.Errorf("platform public key not ready")
	}

	// 生成 AES 密钥
	aesKey := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return "", fmt.Errorf("generate AES key failed: %w", err)
	}
	aesKeyHex := hex.EncodeToString(aesKey)

	// RSA 加密 AES 密钥
	reqParam, err := RSAEncrypt(c.PlatformPubKey, aesKeyHex)
	if err != nil {
		return "", fmt.Errorf("encrypt AES key failed: %w", err)
	}

	// AES 加密请求体
	reqEnc, err := AESEncrypt(body, aesKey)
	if err != nil {
		return "", fmt.Errorf("encrypt body failed: %w", err)
	}

	// 根据 Content-Type 组装
	if strings.Contains(contentType, "application/json") {
		obj := map[string]string{
			"reqEnc":       reqEnc,
			"reqParam":     reqParam,
			"publicKeyPem": c.LocalKeyPair.PublicKey,
		}
		b, _ := json.Marshal(obj)
		return string(b), nil
	}

	return fmt.Sprintf("reqParam=%s&reqEnc=%s&publicKeyPem=%s",
		url.QueryEscape(reqParam),
		url.QueryEscape(reqEnc),
		url.QueryEscape(c.LocalKeyPair.PublicKey)), nil
}

// DecryptResponse 解密响应体
func (c *CryptoContext) DecryptResponse(body string) (string, error) {
	var resp struct {
		Status     string `json:"status"`
		Code       string `json:"code"`
		Msg        string `json:"msg"`
		Data       string `json:"data"`
		EndBAesKey string `json:"EndBAesKey"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	if resp.EndBAesKey == "" {
		// 未加密响应，直接返回
		return body, nil
	}

	if c.LocalKeyPair == nil {
		return "", fmt.Errorf("local key pair missing, cannot decrypt response")
	}

	// RSA 解密 EndBAesKey
	aesKeyHex, err := RSADecrypt(c.LocalKeyPair.PrivateKey, resp.EndBAesKey)
	if err != nil {
		return "", fmt.Errorf("decrypt EndBAesKey failed: %w", err)
	}

	aesKey, err := hex.DecodeString(aesKeyHex)
	if err != nil {
		return "", fmt.Errorf("decode AES key hex failed: %w", err)
	}

	// AES 解密 data
	decryptedData, err := AESDecrypt(resp.Data, aesKey)
	if err != nil {
		return "", fmt.Errorf("decrypt data failed: %w", err)
	}

	// 平台有时会把解密后的 data 再包一层 JSON 字符串，需要二次解析
	var innerStr string
	if err := json.Unmarshal([]byte(decryptedData), &innerStr); err == nil {
		decryptedData = innerStr
	}

	// 如果解密后的 data 是 GBK 编码，尝试转换为 UTF-8
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Data, _, err := transform.Bytes(decoder, []byte(decryptedData))
	if err == nil {
		// 验证转换后的是否是合法 JSON
		var test interface{}
		if json.Unmarshal(utf8Data, &test) == nil {
			decryptedData = string(utf8Data)
		}
	}

	return decryptedData, nil
}

// RefreshPlatformPublicKey 调用平台接口刷新公钥
func RefreshPlatformPublicKey(cookie string) (string, error) {
	req, _ := http.NewRequest("POST", BaseURL+"/openportalsrv/rest/portalmain/main/refreshRSA", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("refreshRSA failed: %w", err)
	}
	defer resp.Body.Close()

	buf := new(strings.Builder)
	io.Copy(buf, resp.Body)
	s := buf.String()

	start := strings.Index(s, `"publicKey":"`)
	if start < 0 {
		return "", fmt.Errorf("publicKey not found in response")
	}
	start += len(`"publicKey":"`)
	end := strings.Index(s[start:], `"`)
	if end < 0 {
		return "", fmt.Errorf("publicKey end not found")
	}
	return strings.ReplaceAll(s[start:start+end], `\n`, "\n"), nil
}
