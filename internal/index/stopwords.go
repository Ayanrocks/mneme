package index

// ProgrammingStopwords contains keywords from top 10 programming languages
// that appear in nearly every code file and don't help BM25 discrimination.
// Sources: Official language specs for Go, Python, JavaScript, Java, C++, Rust, TypeScript, C#, PHP, Ruby
var ProgrammingStopwords = map[string]bool{
	// ===== Go (25 keywords) =====
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,

	// ===== Python =====
	"and": true, "as": true, "assert": true, "async": true, "await": true,
	"class": true, "def": true, "del": true, "elif": true, "except": true,
	"finally": true, "from": true, "global": true, "in": true, "is": true,
	"lambda": true, "nonlocal": true, "not": true, "or": true, "pass": true,
	"raise": true, "try": true, "while": true, "with": true, "yield": true,
	"none": true, "true": true, "false": true,

	// ===== JavaScript/TypeScript =====
	"catch": true, "debugger": true, "delete": true, "do": true, "export": true,
	"extends": true, "function": true, "instanceof": true, "let": true, "new": true,
	"null": true, "super": true, "this": true, "throw": true, "typeof": true,
	"void": true, "declare": true, "infer": true, "keyof": true, "module": true,
	"namespace": true, "never": true, "readonly": true, "unknown": true,

	// ===== Java/C# =====
	"abstract": true, "boolean": true, "byte": true, "char": true, "double": true,
	"enum": true, "final": true, "float": true, "implements": true, "int": true,
	"long": true, "native": true, "private": true, "protected": true, "public": true,
	"short": true, "static": true, "synchronized": true, "throws": true, "transient": true,
	"volatile": true, "sealed": true, "virtual": true, "override": true, "internal": true,
	"using": true, "event": true, "delegate": true, "checked": true, "unchecked": true,

	// ===== C/C++ =====
	"alignas": true, "alignof": true, "asm": true, "auto": true, "constexpr": true,
	"decltype": true, "explicit": true, "extern": true, "friend": true, "inline": true,
	"mutable": true, "noexcept": true, "nullptr": true, "operator": true, "register": true,
	"reinterpret_cast": true, "signed": true, "sizeof": true, "static_assert": true,
	"static_cast": true, "template": true, "thread_local": true, "typedef": true,
	"typeid": true, "typename": true, "union": true, "unsigned": true, "wchar_t": true,
	"dynamic_cast": true, "const_cast": true,

	// ===== Rust =====
	"crate": true, "dyn": true, "fn": true, "impl": true, "loop": true,
	"match": true, "mod": true, "move": true, "mut": true, "pub": true,
	"ref": true, "self": true, "trait": true, "unsafe": true, "use": true,
	"where": true,

	// ===== PHP =====
	"echo": true, "print": true, "require": true, "include": true, "require_once": true,
	"include_once": true, "isset": true, "unset": true, "empty": true, "die": true,
	"exit": true, "eval": true, "list": true, "array": true, "callable": true,
	"iterable": true, "mixed": true,

	// ===== Ruby =====
	"begin": true, "end": true, "ensure": true, "rescue": true, "then": true,
	"unless": true, "until": true, "when": true, "defined": true, "alias": true,
	"redo": true, "retry": true, "undef": true,

	// ===== Common data types (often not useful for search) =====
	"string": true, "integer": true, "number": true, "object": true,

	// ===== Very short tokens (often noise) =====
	"id": true, "ok": true, "err": true, "val": true, "arg": true,
	"args": true, "ctx": true, "fmt": true, "log": true, "msg": true,
	"req": true, "res": true, "str": true, "tmp": true, "buf": true,
}

// IsStopword checks if a token is a programming stopword
func IsStopword(token string) bool {
	return ProgrammingStopwords[token]
}

// FilterStopwords removes programming stopwords from a token slice
func FilterStopwords(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if !IsStopword(token) {
			result = append(result, token)
		}
	}
	return result
}
