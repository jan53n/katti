package main

import (
	"fmt"

	. "jnsn.in/katti"
)

func main() {
	positiveDigit := CharRange('1', '9')
	digit := CharRange('0', '9')
	dot := SingleChar('.')

	numericIdentifier := Alternation(
		SingleChar('0'),
		Sequence(
			positiveDigit,
			Repeat(digit, true),
		),
	)

	nonDigit := Char("[a-zA-Z-]")

	identifierChar := Char("[a-zA-Z0-9-]")

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
					dot,
					Pluck(preReleaseIdentifier),
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
					dot,
					Pluck(buildIdentifier),
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
					SingleChar('-'),
					Pluck(preRelease),
				),
			),
		),
		Bind("build",
			Optional(
				Sequence(
					SingleChar('+'),
					Pluck(build),
				),
			),
		),
	)

	err, result := Parse(semver, "1.0.0-alpha+rr")
	fmt.Printf("%#v, err: %#v\n", err, result)
}
