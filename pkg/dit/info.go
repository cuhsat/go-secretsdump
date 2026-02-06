package dit

import (
	"encoding/hex"
	"fmt"
	"unicode"
)

type uacFlags struct {
	Script                     bool
	AccountDisable             bool
	HomeDirRequired            bool
	Lockout                    bool
	PasswdNotReqd              bool
	EncryptedTextPwdAllowed    bool
	TempDupAccount             bool
	NormalAccount              bool
	InterDomainTrustAcct       bool
	WorkstationTrustAccount    bool
	ServerTrustAccount         bool
	DontExpirePassword         bool
	MNSLogonAccount            bool
	SmartcardRequired          bool
	TrustedForDelegation       bool
	NotDelegated               bool
	UseDESOnly                 bool
	DontPreauth                bool
	PasswordExpired            bool
	TrustedToAuthForDelegation bool
	PartialSecrets             bool
}
type Info struct {
	Username string   `json:"username,omitempty"`
	Lm       []byte   `json:"lm,omitempty"`
	Nt       []byte   `json:"nt,omitempty"`
	Rid      uint32   `json:"rid,omitempty"`
	Enabled  bool     `json:"enabled,omitempty"`
	UAC      uacFlags `json:"uac,omitempty"`
	Supp     SuppInfo `json:"supp,omitempty"`
	History  History  `json:"history,omitempty"`
}

type SuppInfo struct {
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	NotASCII     bool     `json:"not_ascii,omitempty"`
	KerberosKeys []string `json:"kerberos_keys,omitempty"`
}

type History struct {
	Lm [][]byte `json:"lm,omitempty"`
	Nt [][]byte `json:"nt,omitempty"`
}

func (d Info) String() string {
	answer := fmt.Sprintf("%s:%d:%s:%s:::",
		d.Username,
		d.Rid,
		hex.EncodeToString(d.Lm),
		hex.EncodeToString(d.Nt))
	return answer
}

func (d Info) HistoryStrings() []string {
	r := make([]string, 0, len(d.History.Nt))

	for i, v := range d.History.Lm {
		r = append(r, fmt.Sprintf("%s_history%d:%d:%s:%s:::",
			d.Username,
			i,
			d.Rid,
			hex.EncodeToString(v),
			hex.EncodeToString(EmptyNT),
		))
	}

	for i, v := range d.History.Nt {
		r = append(r, fmt.Sprintf("%s_history%d:%d:%s:%s:::",
			d.Username,
			i,
			d.Rid,
			hex.EncodeToString(EmptyLM),
			hex.EncodeToString(v),
		))
	}
	return r
}

// this is a dumb way of doing it,
// but I've had too many rums to think of the actual way
func decodeUAC(val int) uacFlags {
	r := uacFlags{}
	r.Script = val|1 == val
	r.AccountDisable = val|2 == val
	r.HomeDirRequired = val|8 == val
	r.Lockout = val|6 == val
	r.PasswdNotReqd = val|32 == val
	r.EncryptedTextPwdAllowed = val|128 == val
	r.TempDupAccount = val|256 == val
	r.NormalAccount = val|512 == val
	r.InterDomainTrustAcct = val|2048 == val
	r.WorkstationTrustAccount = val|4096 == val
	r.ServerTrustAccount = val|8192 == val
	r.DontExpirePassword = val|65536 == val
	r.MNSLogonAccount = val|131072 == val
	r.SmartcardRequired = val|262144 == val
	r.TrustedForDelegation = val|524288 == val
	r.NotDelegated = val|1048576 == val
	r.UseDESOnly = val|2097152 == val
	r.DontPreauth = val|4194304 == val
	r.PasswordExpired = val|8388608 == val
	r.TrustedToAuthForDelegation = val|16777216 == val
	r.PartialSecrets = val|67108864 == val
	return r
}

// https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
