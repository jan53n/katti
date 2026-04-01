// https://github.com/peggyjs/peggy/blob/main/examples/semver.peggy
package main

import (
	"fmt"

	. "jnsn.in/katti"
)

type ver struct {
	major      string
	minor      string
	patch      string
	preRelease string
	build      string
}

func ParseSemver(raw string) (info ver, err error) {
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

	preRelease := Action(
		Sequence(
			Bind("PR_HEAD", preReleaseIdentifier),
			Bind("PR_TAIL",
				Repeat(
					Sequence(
						skipdot,
						preReleaseIdentifier,
					),
					true,
				),
			),
		),
		func(result *MatchResult) error {
			info.preRelease = result.Bindings.GetString("PR_HEAD") + result.Bindings.GetString("PR_TAIL")
			return nil
		},
	)

	buildIdentifier := Alternation(
		alphanumericIdentifier,
		Repeat(digit, false),
	)

	build := Action(
		Sequence(
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
		),
		func(result *MatchResult) error {
			info.build = result.Bindings.GetString("B_HEAD") + result.Bindings.GetString("B_TAIL")
			return nil
		},
	)

	versionCore := Action(
		Sequence(
			Bind("MAJOR", Join(numericIdentifier)),
			dot,
			Bind("MINOR", Join(numericIdentifier)),
			dot,
			Bind("PATCH", Join(numericIdentifier)),
		),
		func(result *MatchResult) error {
			info.major = result.Bindings.GetString("MAJOR")
			info.minor = result.Bindings.GetString("MINOR")
			info.patch = result.Bindings.GetString("PATCH")
			return nil
		},
	)

	semver := Sequence(
		versionCore,
		Optional(
			Sequence(
				Skip(Char('-')),
				preRelease,
			),
		),
		Optional(
			Sequence(
				Skip(Char('+')),
				build,
			),
		),
	)

	if _, err := Parse(semver, raw); err != nil {
		return info, err
	}

	return info, err
}

func main() {
	if info, err := ParseSemver("1.2.3-alpha+rr"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%#v\n", info)
	}
}
