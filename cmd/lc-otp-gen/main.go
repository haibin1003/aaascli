// lc-otp-gen - 命令行版 Google Authenticator
// 用于测试 OTP 功能，生成 TOTP 验证码
package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

// OTPConfig 存储的 OTP 配置
type OTPConfig struct {
	Account string `json:"account"`
	Secret  string `json:"secret"`
	Issuer  string `json:"issuer"`
}

// Configs 配置集合
type Configs struct {
	OTPs []OTPConfig `json:"otps"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "lc-otp-gen",
		Short: "命令行版 Google Authenticator",
		Long:  `用于生成 TOTP 验证码，配合 lc otp 命令使用`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ~/.lc-otp-gen/config.json)")

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(codeCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(removeCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// addCmd 添加新的 OTP 配置
func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [account] [secret]",
		Short: "添加 OTP 配置",
		Long: `添加新的 OTP 配置

示例:
  lc-otp-gen add weibaohui@hq.cmcc HAEXHXIW6QQVFLUPYOVIGQTY7MYPZMKK
  lc-otp-gen add myaccount "JBSW Y3DP EHPK 3PXP"`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			account := args[0]
			secret := strings.ToUpper(strings.ReplaceAll(args[1], " ", ""))

			configs, err := loadConfigs()
			if err != nil {
				configs = &Configs{OTPs: []OTPConfig{}}
			}

			// 检查是否已存在
			for i, cfg := range configs.OTPs {
				if cfg.Account == account {
					// 更新
					configs.OTPs[i].Secret = secret
					if err := saveConfigs(configs); err != nil {
						fmt.Printf("保存失败: %v\n", err)
						return
					}
					fmt.Printf("✅ 已更新 OTP 配置: %s\n", account)
					return
				}
			}

			// 添加新配置
			configs.OTPs = append(configs.OTPs, OTPConfig{
				Account: account,
				Secret:  secret,
				Issuer:  "灵畿CLI",
			})

			if err := saveConfigs(configs); err != nil {
				fmt.Printf("保存失败: %v\n", err)
				return
			}

			fmt.Printf("✅ 已添加 OTP 配置: %s\n", account)
			fmt.Println("\n现在可以运行: lc-otp-gen code", account)
		},
	}
}

// codeCmd 生成验证码
func codeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "code [account]",
		Short: "生成当前验证码",
		Long: `生成指定账户的当前 TOTP 验证码

如果不指定账户，且只有一个配置，则使用该配置

示例:
  lc-otp-gen code                    # 使用唯一配置
  lc-otp-gen code weibaohui@hq.cmcc  # 使用指定账户`,
		Run: func(cmd *cobra.Command, args []string) {
			configs, err := loadConfigs()
			if err != nil || len(configs.OTPs) == 0 {
				fmt.Println("❌ 没有 OTP 配置，请先运行: lc-otp-gen add [account] [secret]")
				return
			}

			var config OTPConfig

			if len(args) == 0 {
				if len(configs.OTPs) == 1 {
					config = configs.OTPs[0]
				} else {
					fmt.Println("❌ 有多个配置，请指定账户:")
					for _, cfg := range configs.OTPs {
						fmt.Printf("  - %s\n", cfg.Account)
					}
					return
				}
			} else {
				account := args[0]
				found := false
				for _, cfg := range configs.OTPs {
					if cfg.Account == account {
						config = cfg
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("❌ 未找到配置: %s\n", account)
					return
				}
			}

			// 生成验证码
			code, remaining := generateTOTP(config.Secret)

			fmt.Printf("\n🔐 账户: %s\n", config.Account)
			fmt.Printf("🔢 验证码: %s\n", code)
			fmt.Printf("⏱️  剩余: %d 秒\n\n", remaining)
		},
	}
}

// listCmd 列出所有配置
func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有 OTP 配置",
		Run: func(cmd *cobra.Command, args []string) {
			configs, err := loadConfigs()
			if err != nil || len(configs.OTPs) == 0 {
				fmt.Println("没有 OTP 配置")
				return
			}

			fmt.Println("\n📋 OTP 配置列表:")
			fmt.Println("────────────────────────────────────────")
			for i, cfg := range configs.OTPs {
				code, remaining := generateTOTP(cfg.Secret)
				fmt.Printf("%d. %s\n", i+1, cfg.Account)
				fmt.Printf("   当前验证码: %s (剩余 %d 秒)\n", code, remaining)
				fmt.Printf("   密钥: %s\n", formatSecret(cfg.Secret))
				fmt.Println()
			}
		},
	}
}

// removeCmd 删除配置
func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [account]",
		Short: "删除 OTP 配置",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			account := args[0]

			configs, err := loadConfigs()
			if err != nil {
				fmt.Printf("读取配置失败: %v\n", err)
				return
			}

			newOTPs := []OTPConfig{}
			found := false
			for _, cfg := range configs.OTPs {
				if cfg.Account == account {
					found = true
					continue
				}
				newOTPs = append(newOTPs, cfg)
			}

			if !found {
				fmt.Printf("❌ 未找到配置: %s\n", account)
				return
			}

			configs.OTPs = newOTPs
			if err := saveConfigs(configs); err != nil {
				fmt.Printf("保存失败: %v\n", err)
				return
			}

			fmt.Printf("✅ 已删除配置: %s\n", account)
		},
	}
}

// generateTOTP 生成 TOTP 验证码
func generateTOTP(secret string) (string, int) {
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "ERROR", 0
	}

	now := time.Now()
	counter := uint64(math.Floor(float64(now.Unix()) / 30))

	mac := hmac.New(sha1.New, key)
	binary.Write(mac, binary.BigEndian, counter)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF
	code = code % 1000000

	remaining := 30 - (now.Unix() % 30)
	return fmt.Sprintf("%06d", code), int(remaining)
}

// formatSecret 格式化密钥显示
func formatSecret(secret string) string {
	var result strings.Builder
	for i, c := range secret {
		if i > 0 && i%4 == 0 {
			result.WriteByte(' ')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".lc-otp-gen.json"
	}

	configDir := filepath.Join(home, ".lc-otp-gen")
	os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, "config.json")
}

// loadConfigs 加载配置
func loadConfigs() (*Configs, error) {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var configs Configs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	return &configs, nil
}

// saveConfigs 保存配置
func saveConfigs(configs *Configs) error {
	path := getConfigPath()

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
