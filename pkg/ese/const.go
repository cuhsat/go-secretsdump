package ese

const pageSize = 8192

const FLAGS_LEAF = 2
const FLAGS_SPACE_TREE = 0x20
const FLAGS_INDEX = 0x40
const FLAGS_LONG_VALUE = 0x80

const TAG_COMMON = 0x4

const CATALOG_PAGE_NUMBER = 4

const CATALOG_TYPE_TABLE = 1
const CATALOG_TYPE_COLUMN = 2
const CATALOG_TYPE_INDEX = 3
const CATALOG_TYPE_LONG_VALUE = 4
const CATALOG_TYPE_CALLBACK = 5

const JET_coltypNil = 0
const JET_coltypBit = 1
const JET_coltypUnsignedByte = 2
const JET_coltypShort = 3
const JET_coltypLong = 4
const JET_coltypCurrency = 5
const JET_coltypIEEESingle = 6
const JET_coltypIEEEDouble = 7
const JET_coltypDateTime = 8
const JET_coltypBinary = 9
const JET_coltypText = 10
const JET_coltypLongBinary = 11
const JET_coltypLongText = 12
const JET_coltypSLV = 13
const JET_coltypUnsignedLong = 14
const JET_coltypLongLong = 15
const JET_coltypGUID = 16
const JET_coltypUnsignedShort = 17
const JET_coltypMax = 18

const TAGGED_DATA_TYPE_COMPRESSED = 2
const TAGGED_DATA_TYPE_MULTI_VALUE = 8
