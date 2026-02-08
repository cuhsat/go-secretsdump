package ese

const pageSize = 8192

const (
	FlagsLeaf      = 2
	FlagsSpaceTree = 0x20
	FlagsIndex     = 0x40
	FlagsLongValue = 0x80
)

const (
	TagCommon = 0x4
)

const (
	CatalogPageNumber = 4
)

const (
	CatalogTypeTable     = 1
	CatalogTypeColumn    = 2
	CatalogTypeIndex     = 3
	CatalogTypeLongValue = 4
	CatalogTypeCallback  = 5
)

const (
	JetColtypnil           = 0
	JetColtypbit           = 1
	JetColtypunsignedbyte  = 2
	JetColtypshort         = 3
	JetColtyplong          = 4
	JetColtypcurrency      = 5
	JetColtypieeesingle    = 6
	JetColtypieeedouble    = 7
	JetColtypdatetime      = 8
	JetColtypbinary        = 9
	JetColtyptext          = 10
	JetColtyplongbinary    = 11
	JetColtyplongtext      = 12
	JetColtypslv           = 13
	JetColtypunsignedlong  = 14
	JetColtyplonglong      = 15
	JetColtypguid          = 16
	JetColtypunsignedshort = 17
	JetColtypmax           = 18
)

const (
	TaggedDataTypeCompressed = 2
	TaggedDataTypeMultiValue = 8
)
