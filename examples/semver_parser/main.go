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
		Bind("pr_head", preReleaseIdentifier),
		Bind("pr_tail",
			Repeat(
				Sequence(
					skipdot,
					preReleaseIdentifier,
				),
				true,
			),
		),
	)

	buildIdentifier := Alternation(
		alphanumericIdentifier,
		Repeat(digit, false),
	)

	build := Sequence(
		Bind("b_head", buildIdentifier),
		Bind("b_tail",
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
		Bind("major", numericIdentifier),
		dot,
		Bind("minor", numericIdentifier),
		dot,
		Bind("patch", numericIdentifier),
	)

	semver := Sequence(
		Bind("versionCore", versionCore),
		Bind("pre",
			Optional(
				Sequence(
					Skip(Char('-')),
					preRelease,
				),
			),
		),
		Bind("build",
			Optional(
				Sequence(
					Skip(Char('+')),
					build,
				),
			),
		),
	)

	result, err := Parse(semver, "1.0.0-alpha+rr")
	fmt.Printf("%#v, err: %#v\n", err, result)
}
