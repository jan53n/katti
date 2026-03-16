// https://github.com/peggyjs/peggy/blob/main/examples/semver.peggy
package main

import (
	"fmt"

	. "jnsn.in/katti"
)

func main() {
	positiveDigit := CharIn('1', '9')
	digit := CharIn('0', '9')
	dot := Char('.')
	skipdot := Skip(dot)

	numericIdentifier := Alternation(
		Char('0'),
		Sequence(
			positiveDigit,
			Repeat(digit, true),
		),
	)

	nonDigit := Alternation(
		CharIn('a', 'z'),
		CharIn('A', 'Z'),
		CharIn('0', '9'),
	)

	identifierChar := Alternation(
		CharIn('a', 'z'),
		CharIn('A', 'Z'),
		CharIn('0', '9'),
		Char('-'),
	)

	alphanumericIdentifier := Sequence(
		Repeat(digit, true),
		nonDigit,
		Repeat(identifierChar, true),
	)

	preReleaseIdentifier := Alternation(
		alphanumericIdentifier,
		numericIdentifier,
	)

	preRelease := Sequence(
		Bind("PR_HEAD", Join(preReleaseIdentifier)),
		Bind("PR_TAIL",
			Join(
				Repeat(
					Sequence(
						skipdot,
						preReleaseIdentifier,
					),
					true,
				),
			),
		),
	)

	buildIdentifier := Alternation(
		alphanumericIdentifier,
		Repeat(digit, false),
	)

	build := Sequence(
		Bind("B_HEAD", buildIdentifier),
		Bind("B_TAIL",
			Repeat(
				Sequence(
					skipdot,
					buildIdentifier,
				),
				true,
			),
		),
	)

	versionCore := Sequence(
		Bind("MAJOR", Join(numericIdentifier)),
		dot,
		Bind("MINOR", Join(numericIdentifier)),
		dot,
		Bind("PATCH", Join(numericIdentifier)),
	)

	semver := Sequence(
		Bind("VERSION_CORE", Join(versionCore)),
		Bind("PRE",
			Optional(
				Sequence(
					Skip(Char('-')),
					preRelease,
				),
			),
		),
		Bind("BUILD",
			Join(
				Optional(
					Sequence(
						Skip(Char('+')),
						build,
					),
				),
			),
		),
	)

	result, err := Parse(semver, "1.0.0-alpha+rr")
	fmt.Printf("result:%#v\nerr: %#v\n", result, err)
}
