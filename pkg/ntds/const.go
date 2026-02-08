package ntds

const (
	nobjectSid               = "ATTr589970"
	nuserAccountControl      = "ATTj589832"
	nsAMAccountName          = "ATTm590045"
	nsAMAccountType          = "ATTj590126"
	nuserPrincipalName       = "ATTm590480"
	nunicodePwd              = "ATTk589914"
	ndBCSPwd                 = "ATTk589879"
	nntPwdHistory            = "ATTk589918"
	nlmPwdHistory            = "ATTk589984"
	npekList                 = "ATTk590689"
	nsupplementalCredentials = "ATTk589949"
)

var kerbkeytype = map[uint32]string{
	1:          "des-cbc-crc",
	3:          "des-cbc-md5",
	17:         "aes128-cts-hmac-sha1-96",
	18:         "aes256-cts-hmac-sha1-96",
	0xffffff74: "rc4-hmac",
}

var accTypes = map[int32]string{
	0x30000000: "SAM_NORMAL_USER_ACCOUNT",
	0x30000001: "SAM_MACHINE_ACCOUNT",
	0x30000002: "SAM_TRUST_ACCOUNT",
}

var EmptyNT = []byte{0x31, 0xd6, 0xcf, 0xe0, 0xd1, 0x6a, 0xe9, 0x31, 0xb7, 0x3c, 0x59, 0xd7, 0xe0, 0xc0, 0x89, 0xc0}
var EmptyLM = []byte{0xaa, 0xd3, 0xb4, 0x35, 0xb5, 0x14, 0x04, 0xee, 0xaa, 0xd3, 0xb4, 0x35, 0xb5, 0x14, 0x04, 0xee}
