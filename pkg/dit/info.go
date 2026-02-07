package dit

import (
	"encoding/hex"
	"fmt"
)

type Credentials struct {
	Username     string       `json:"username,omitempty"`
	Lm           []byte       `json:"lm,omitempty"`
	Nt           []byte       `json:"nt,omitempty"`
	Rid          uint32       `json:"rid,omitempty"`
	Enabled      bool         `json:"enabled,omitempty"`
	Uac          Flags        `json:"uac,omitempty"`
	Supplemental Supplemental `json:"supplemental,omitempty"`
	History      History      `json:"history,omitempty"`
}

type Flags struct {
	Script                       bool `json:"script,omitempty"`
	AccountDisable               bool `json:"account_disable,omitempty"`
	HomeDirRequired              bool `json:"home_dir_required,omitempty"`
	Lockout                      bool `json:"lockout,omitempty"`
	PasswordNotRequired          bool `json:"password_not_required,omitempty"`
	EncryptedTextPasswordAllowed bool `json:"encrypted_text_password_allowed,omitempty"`
	TemporaryDupAccount          bool `json:"temporary_dup_account,omitempty"`
	NormalAccount                bool `json:"normal_account,omitempty"`
	InterDomainTrustAccount      bool `json:"inter_domain_trust_account,omitempty"`
	WorkstationTrustAccount      bool `json:"workstation_trust_account,omitempty"`
	ServerTrustAccount           bool `json:"server_trust_account,omitempty"`
	DontExpirePassword           bool `json:"dont_expire_password,omitempty"`
	MNSLogonAccount              bool `json:"mns_logon_account,omitempty"`
	SmartCardRequired            bool `json:"smart_card_required,omitempty"`
	TrustedForDelegation         bool `json:"trusted_for_delegation,omitempty"`
	NotDelegated                 bool `json:"not_delegated,omitempty"`
	UseDESOnly                   bool `json:"use_des_only,omitempty"`
	DontPreAuth                  bool `json:"dont_pre_auth,omitempty"`
	PasswordExpired              bool `json:"password_expired,omitempty"`
	TrustedToAuthForDelegation   bool `json:"trusted_to_auth_for_delegation,omitempty"`
	PartialSecrets               bool `json:"partial_secrets,omitempty"`
}

type Supplemental struct {
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	NotASCII     bool     `json:"not_ascii,omitempty"`
	KerberosKeys []string `json:"kerberos_keys,omitempty"`
}

type History struct {
	Lm [][]byte `json:"lm,omitempty"`
	Nt [][]byte `json:"nt,omitempty"`
}

func (c Credentials) String() string {
	answer := fmt.Sprintf("%s:%d:%s:%s:::",
		c.Username,
		c.Rid,
		hex.EncodeToString(c.Lm),
		hex.EncodeToString(c.Nt))
	return answer
}

func (c Credentials) GetHistory() []string {
	h := make([]string, 0, len(c.History.Nt))

	for i, v := range c.History.Lm {
		h = append(h, fmt.Sprintf("%s_history%d:%d:%s:%s:::",
			c.Username,
			i,
			c.Rid,
			hex.EncodeToString(v),
			hex.EncodeToString(EmptyNT),
		))
	}

	for i, v := range c.History.Nt {
		h = append(h, fmt.Sprintf("%s_history%d:%d:%s:%s:::",
			c.Username,
			i,
			c.Rid,
			hex.EncodeToString(EmptyLM),
			hex.EncodeToString(v),
		))
	}

	return h
}

func decodeUAC(v int) Flags {
	return Flags{
		Script:                       v|1 == v,
		AccountDisable:               v|2 == v,
		HomeDirRequired:              v|8 == v,
		Lockout:                      v|6 == v,
		PasswordNotRequired:          v|32 == v,
		EncryptedTextPasswordAllowed: v|128 == v,
		TemporaryDupAccount:          v|256 == v,
		NormalAccount:                v|512 == v,
		InterDomainTrustAccount:      v|2048 == v,
		WorkstationTrustAccount:      v|4096 == v,
		ServerTrustAccount:           v|8192 == v,
		DontExpirePassword:           v|65536 == v,
		MNSLogonAccount:              v|131072 == v,
		SmartCardRequired:            v|262144 == v,
		TrustedForDelegation:         v|524288 == v,
		NotDelegated:                 v|1048576 == v,
		UseDESOnly:                   v|2097152 == v,
		DontPreAuth:                  v|4194304 == v,
		PasswordExpired:              v|8388608 == v,
		TrustedToAuthForDelegation:   v|16777216 == v,
		PartialSecrets:               v|67108864 == v,
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7F {
			return false
		}
	}
	return true
}
