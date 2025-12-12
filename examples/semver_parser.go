package main

import (
	"fmt"

	. "jnsn.in/katti"
)

func main() {
	positiveDigit := CharIn('1', '9')
	digit := CharIn('0', '9')
	dot := SingleChar('.')

	numericIdentifier := Alternation(
		SingleChar('0'),
		Sequence(
			positiveDigit,
			Repeat(digit, true),
		),
	)

	nonDigit := Char(
		[]CharRange{
			{Start: 'a', End: 'z'},
			{Start: 'A', End: 'Z'},
			{Start: '-', End: '-'},
		},
		false,
	)

	identifierChar := Char(
		[]CharRange{
			{Start: 'a', End: 'z'},
			{Start: 'A', End: 'Z'},
			{Start: '0', End: '9'},
			{Start: '-', End: '-'},
		},
		false,
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

	result, err := Parse(semver, "1.0.0-alpha+rr")
	fmt.Printf("%#v, err: %#v\n", err, result)
}
