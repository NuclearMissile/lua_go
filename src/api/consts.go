package api

const (
	LUA_MINSATCK            = 20
	LUAI_MAXSTACK           = 1000000
	LUA_REGISTRYINDEX       = -LUAI_MAXSTACK - 1000
	LUA_RIDX_GLOBALS  int64 = 2
)

// basic types
const (
	LUA_TNONE = iota - 1 // -1, 0, 1, 2...
	LUA_TNIL
	LUA_TBOOLEAN
	LUA_TLIGHTUSERDATA
	LUA_TNUMBER
	LUA_TSTRING
	LUA_TTABLE
	LUA_TFUNCTION
	LUA_TUSERDATA
	LUA_TTHREAD
)

// lua-5.3.4/src/lobject.h
/* type variants */
const (
	LUA_TNUMFLT = LUA_TNUMBER | (0 << 4)   // float numbers
	LUA_TNUMINT = LUA_TNUMBER | (1 << 4)   // integer numbers
	LUA_TSHRSTR = LUA_TSTRING | (0 << 4)   // short strings
	LUA_TLNGSTR = LUA_TSTRING | (1 << 4)   // long strings
	LUA_TLCL    = LUA_TFUNCTION | (0 << 4) // Lua closure
	LUA_TLGF    = LUA_TFUNCTION | (1 << 4) // light Go function
	LUA_TGCL    = LUA_TFUNCTION | (2 << 4) // Go closure
)

/* arithmetic functions */
const (
	LUA_OPADD  = iota // +
	LUA_OPSUB         // -
	LUA_OPMUL         // *
	LUA_OPMOD         // %
	LUA_OPPOW         // ^
	LUA_OPDIV         // /
	LUA_OPIDIV        // //
	LUA_OPBAND        // &
	LUA_OPBOR         // |
	LUA_OPBXOR        // ~
	LUA_OPSHL         // <<
	LUA_OPSHR         // >>
	LUA_OPUNM         // -
	LUA_OPBNOT        // ~
)

// function call status
const (
	LUA_OK = iota
	LUA_YIELD
	LUA_ERRRUN
	LUA_ERRSYNTAX
	LUA_ERRMEM
	LUA_ERRGCMM
	LUA_ERRERR
	LUA_ERRFILE
)

// comparison functions
const (
	LUA_OPEQ = iota
	// <
	LUA_OPLT
	// <=
	LUA_OPLE
)
