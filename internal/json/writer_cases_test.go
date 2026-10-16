// SPDX-License-Identifier: MIT
// Provenance: original

package json

// Texts produced by the upstream library (plantuml-mit 1.2026.8).
const (
	// Json.value(all) where all is U+0000..U+007F followed by U+0080, U+00FF,
	// U+2027, U+2028, U+2029, U+202A, U+1F600 and U+FEFF.
	escapeAllWant = "\"\\u0000\\u0001\\u0002\\u0003\\u0004\\u0005\\u0006\\u0007\\u0008\\t\\n\\u000b\\u000c\\r\\u000e\\u000f\\u0010\\u0011\\u0012\\u0013\\u0014\\u0015\\u0016\\u0017\\u0018\\u0019\\u001a\\u001b\\u001c\\u001d\\u001e\\u001f !\\\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\x7f\u0080\u00ff\u2027\\u2028\\u2029\u202a\U0001f600\ufeff\""
	// new JsonObject().add(name, "v") where name is k, a quote, U+2028 and U+0001.
	escapeMemberNameWant = "{\"k\\\"\\u2028\\u0001\":\"v\"}"
	// {"a":[1,{}],"b":{"c":"d"}} with indentWithSpaces(0) and indentWithSpaces(4).
	indent0Want = "{\n\"a\": [\n1,\n{\n\n}\n],\n\"b\": {\n\"c\": \"d\"\n}\n}"
	indent4Want = "{\n    \"a\": [\n        1,\n        {\n            \n        }\n    ],\n    \"b\": {\n        \"c\": \"d\"\n    }\n}"
)
