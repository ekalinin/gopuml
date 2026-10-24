// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/Transform.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

// Transformations on dictionary words.

const numTransforms = 121

var (
	transforms        [numTransforms * 3]int
	prefixSuffix      [217]byte
	prefixSuffixHeads [51]int
)

// Bundle of 0-terminated strings. The packed tables are Java strings of
// characters below 0x100, here bytes: the \xC2\xA0 of prefixSuffixSrc is
// the two Java chars \u00C2\u00A0.
const (
	prefixSuffixSrc = "# #s #, #e #.# the #.com/#\xC2\xA0# of # and" +
		" # in # to #\"#\">#\n#]# for # a # that #. # with #'# from # by #. The # on # as # is #ing" +
		" #\n\t#:#ed #(# at #ly #=\"# of the #. This #,# not #er #al #='#ful #ive #less #est #ize #" + "ous #"
	transformsSrc = "     !! ! ,  *!  &!  \" !  ) *   * -  ! # !  #!*!  " +
		"+  ,$ !  -  %  .  / #   0  1 .  \"   2  3!*   4%  ! # /   5  6  7  8 0  1 &   $   9 +   : " +
		" ;  < '  !=  >  ?! 4  @ 4  2  &   A *# (   B  C& ) %  ) !*# *-% A +! *.  D! %'  & E *6  F " +
		" G% ! *A *%  H! D  I!+!  J!+   K +- *4! A  L!*4  M  N +6  O!*% +.! K *G  P +%(  ! G *D +D " +
		" Q +# *K!*G!+D!+# +G +A +4!+% +K!+4!*D!+K!*K"
)

func unpackTransforms(prefixSuffix []byte, prefixSuffixHeads []int, transforms []int,
	prefixSuffixSrc, transformsSrc string) {
	n := len(prefixSuffixSrc)
	index := 1
	for i := range n {
		c := prefixSuffixSrc[i]
		prefixSuffix[i] = c
		if c == 35 { // == #
			prefixSuffixHeads[index] = i + 1
			index++
			prefixSuffix[i] = 0
		}
	}

	for i := range numTransforms * 3 {
		transforms[i] = int(transformsSrc[i]) - 32
	}
}

func init() {
	unpackTransforms(prefixSuffix[:], prefixSuffixHeads[:], transforms[:], prefixSuffixSrc, transformsSrc)
}

// transformDictionaryWord writes the word of length bytes at
// data[wordOffset:], transformed by the transform transformIndex, to
// dst[dstOffset:] and returns the number of bytes written.
func transformDictionaryWord(dst []byte, dstOffset int, data []byte, wordOffset, length, transformIndex int) int {
	offset := dstOffset
	transformOffset := 3 * transformIndex
	transformPrefix := prefixSuffixHeads[transforms[transformOffset]]
	transformType := transforms[transformOffset+1]
	transformSuffix := prefixSuffixHeads[transforms[transformOffset+2]]

	// Copy prefix.
	for prefixSuffix[transformPrefix] != 0 {
		dst[offset] = prefixSuffix[transformPrefix]
		offset++
		transformPrefix++
	}

	// Copy trimmed word.
	omitFirst := 0
	if transformType >= 12 {
		omitFirst = transformType - 11
	}
	if omitFirst > length {
		omitFirst = length
	}
	wordOffset += omitFirst
	length -= omitFirst
	if transformType <= 9 { // Omit last.
		length -= transformType
	}
	i := length
	for i > 0 {
		dst[offset] = data[wordOffset]
		offset++
		wordOffset++
		i--
	}

	// Ferment.
	if transformType == 11 || transformType == 10 {
		uppercaseOffset := offset - length
		if transformType == 10 {
			length = 1
		}
		for length > 0 {
			tmp := int(dst[uppercaseOffset])
			if tmp < 0xc0 {
				if tmp >= 97 && tmp <= 122 { // in [a..z] range
					dst[uppercaseOffset] ^= 32
				}
				uppercaseOffset += 1
				length -= 1
			} else if tmp < 0xe0 {
				dst[uppercaseOffset+1] ^= 32
				uppercaseOffset += 2
				length -= 2
			} else {
				dst[uppercaseOffset+2] ^= 5
				uppercaseOffset += 3
				length -= 3
			}
		}
	}

	// Copy suffix.
	for prefixSuffix[transformSuffix] != 0 {
		dst[offset] = prefixSuffix[transformSuffix]
		offset++
		transformSuffix++
	}

	return offset - dstOffset
}
