package si

import (
	"errors"
	"math/bits"
	"strconv"
	"unicode/utf8"
)

const (
	maxunit = 200
)

// NewDimension creates a new dimension from the given exponents.
func NewDimension(length, mass, time, temperature, electricCurrent, luminosity, amount int) (Dimension, error) {
	if isDimOOB(length) || isDimOOB(mass) || isDimOOB(time) ||
		isDimOOB(temperature) || isDimOOB(electricCurrent) ||
		isDimOOB(luminosity) || isDimOOB(amount) {
		return Dimension{}, errors.New("overflow dimension storage")
	}
	return Dimension{
		dims: [7]int16{
			int16(length), int16(mass),
			int16(time), int16(temperature),
			int16(electricCurrent), int16(luminosity),
			int16(amount),
		},
	}, nil
}

func isDimOOB(dim int) bool {
	return dim > maxunit || dim < -maxunit
}

// Dimension represents the dimensions of a physical quantity.
type Dimension struct {
	// dims contains 7 int4's representing the exponent of primitive dimensions:
	//  0. Distance dimension (L)
	//  1. Mass dimension (M)
	//  2. Time dimension (T)
	//  3. Temperature dimension (K)
	//  4. Electric current dimension (I)
	//  5. Luminous intensity dimension (J)
	//  6. Amount or quantity dimension (N). i.e: moles, particles, electric pulses, etc.
	// Since these are int4s they are in the range of -8 to 7. Last 4 bits of fourth byte are unused.
	dims [7]int16
}

const negexp = '⁻'

var exprune = [...]rune{
	0: '⁰',
	1: '¹',
	2: '²',
	3: '³',
	4: '⁴',
	5: '⁵',
	6: '⁶',
	7: '⁷',
	8: '⁸',
	9: '⁹',
}

func (d Dimension) String() string {
	if d == (Dimension{}) {
		return ""
	}
	s := make([]byte, 0, 8)
	return string(d.appendf(s))
}

func (d Dimension) appendf(b []byte) []byte {
	app := func(b []byte, char byte, dim int) []byte {
		if dim == 0 {
			return b
		}
		b = append(b, char)
		if dim == 1 {
			return b
		}

		var buf [20]byte
		numbuf := strconv.AppendInt(buf[:0], int64(dim), 10)
		if numbuf[0] == '-' {
			b = utf8.AppendRune(b, negexp)
			numbuf = numbuf[1:]
		}
		for i := 0; i < len(numbuf); i++ {
			offset := numbuf[i] - '0'
			if offset > 9 {
				panic("invalid char")
			}
			b = utf8.AppendRune(b, exprune[offset])
		}
		return b
	}
	b = app(b, 'L', d.ExpLength())
	b = app(b, 'M', d.ExpMass())
	b = app(b, 'T', d.ExpTime())
	b = app(b, 'K', d.ExpTemperature())
	b = app(b, 'I', d.ExpCurrent())
	b = app(b, 'J', d.ExpLuminous())
	b = app(b, 'N', d.ExpAmount())
	return b
}

func (d Dimension) ExpLength() int      { return int(d.dims[0]) }
func (d Dimension) ExpMass() int        { return int(d.dims[1]) }
func (d Dimension) ExpTime() int        { return int(d.dims[2]) }
func (d Dimension) ExpTemperature() int { return int(d.dims[3]) }
func (d Dimension) ExpCurrent() int     { return int(d.dims[4]) }
func (d Dimension) ExpLuminous() int    { return int(d.dims[5]) }
func (d Dimension) ExpAmount() int      { return int(d.dims[6]) }

// func (d Dimension) ExpLength() int      { return i4ToInt(d.dims[0] & 0xf) }
// func (d Dimension) ExpMass() int        { return i4ToInt(d.dims[0] >> 4) }
// func (d Dimension) ExpTime() int        { return i4ToInt(d.dims[1] & 0xf) }
// func (d Dimension) ExpTemperature() int { return i4ToInt(d.dims[1] >> 4) }
// func (d Dimension) ExpCurrent() int     { return i4ToInt(d.dims[2] & 0xf) }
// func (d Dimension) ExpLuminous() int    { return i4ToInt(d.dims[2] >> 4) }
// func (d Dimension) ExpAmount() int      { return i4ToInt(d.dims[3] & 0xf) }

// Inv inverts the dimension by multiplying all dimension exponents by -1.
func (d Dimension) Inv() Dimension {
	const negativeBits = 0b1000_1000
	inv := d
	inv.dims[0] ^= negativeBits
	inv.dims[1] ^= negativeBits
	inv.dims[2] ^= negativeBits
	inv.dims[3] ^= negativeBits
	return inv
}

func MulDim(a, b Dimension) (Dimension, error) {
	L := a.ExpLength() + b.ExpLength()
	M := a.ExpMass() + b.ExpMass()
	T := a.ExpTime() + b.ExpTime()
	K := a.ExpTemperature() + b.ExpTemperature()
	I := a.ExpCurrent() + b.ExpCurrent()
	J := a.ExpLuminous() + b.ExpLuminous()
	N := a.ExpAmount() + b.ExpAmount()
	return NewDimension(L, M, T, K, I, J, N)
}

func DivDim(a, b Dimension) (Dimension, error) {
	return MulDim(a, b.Inv())
}

const negmask = 1 << 3

func intToI4(c int) byte {
	if c < 0 {
		c |= negmask
	}
	return byte(c) & 0xf
}

// i4ToInt converts lower 4 bits of a byte to a signed 4 bit integer.
func i4ToInt(c byte) (v int) {
	v = int(c & 0xf)
	if v&negmask != 0 {
		v &= 0xf &^ negmask
		v = -v
	}
	return v
}

type Prefix int8

const (
	PrefixAtto Prefix = -18 + iota*3
	PrefixFemto
	PrefixPico
	PrefixNano
	PrefixMicro
	PrefixMilli
	PrefixNone
	PrefixKilo
	PrefixMega
	PrefixGiga
	PrefixTera
	PrefixExa
)

// IsValid checks if the prefix is one of the supported standard SI prefixes or the zero base prefix.
func (p Prefix) IsValid() bool {
	return p == PrefixNone || p.Character() != ' '
}

// String returns a human readable representation of the Prefix of string type.
// Returns a error message string if Prefix is undefined. Guarateed to return non-zero string.
func (p Prefix) String() string {
	if p == PrefixMicro {
		return "μ"
	}
	const pfxTable = "a!!f!!p!!n!!u!!m!! !!k!!M!!G!!T!!E"
	offset := int(p - PrefixAtto)
	if offset < 0 || offset >= len(pfxTable) || pfxTable[offset] == '!' {
		return "<si!invalid Prefix>"
	}
	return pfxTable[offset : offset+1]
}

// Character returns the single character SI representation of the unit prefix.
// If not representable or invalid returns space caracter ' '.
func (p Prefix) Character() (s rune) {
	if p == PrefixMicro {
		return 'μ'
	}
	s = rune(p.String()[0])
	if s == '<' {
		s = ' '
	}
	return s
}

// fixed point representation integer supported by this package.
type fixed interface {
	~int64
}

func formatAppend[T fixed](b []byte, value T, base Prefix, fmt byte, prec int) []byte {
	switch {
	case fmt != 'f':
		return append(b, "<si!INVALID FMT>"...)
	case prec < 0:
		return append(b, "<si!NEGATIVE PREC>"...)
	case value == 0:
		return append(b, '0')
	}
	if !base.IsValid() {
		return append(b, "<si!BAD BASE>"...)
	}
	isNegative := value < 0
	if isNegative {
		value = -value
	}

	log10 := ilog10(int64(value))

	/* Description of algorithm:
	First we

	*/

	_ = log10

	b = strconv.AppendFloat(b, float64(value), fmt, prec, 32)
	b = append(b, string(base.Character())...)
	return b
}

var iLogTable = [...]int64{
	1,
	10,
	100,
	1_000,
	10_000,
	100_000,
	1_000_000,
	10_000_000,
	100_000_000,
	1_000_000_000,
	10_000_000_000,
	100_000_000_000,
	1_000_000_000_000,
	10_000_000_000_000,
	100_000_000_000_000,
	1_000_000_000_000_000,
	10_000_000_000_000_000,
	100_000_000_000_000_000,
	1_000_000_000_000_000_000,
}

// ilog10 returns the integer logarithm base 10 of v, which
// can be interpreted as the quanity of digits in the number in base 10.
func ilog10(v int64) int {
	for i, l := range iLogTable {
		if v < l {
			return i
		}
	}
	return len(iLogTable)
}

func formatFixed[T fixed](b []byte, value T, decimal, prec int) []byte {
	const zerostr = ".0000000000000000"
	if decimal > 16 || decimal < -16 {
		return append(b, "<si!DECIMAL OOB>"...)
	} else if value == 0 {
		return append(b, '0')
	}
	isNegative := value < 0
	if isNegative {
		value = abs(value)
	}

	// log10 is amount of significant digits in the
	// value of our fixed point representation.
	log10 := ilog10(int64(value))

	// We calculate amount of digits in front and behind the decimal.
	backDigits := log10 - decimal
	// Now we calculate the fractional part.
	var whole, frac T
	if backDigits > 0 {
		divisor := T(iLogTable[backDigits])
		frac = value % divisor
		if value >= divisor {
			// Avoid division if possible.
			whole /= divisor
		} else if value == 0 && frac == 0 {
			return append(b, '0')
		}
	} else if decimal > 0 {
		hi, lo := bits.Mul64(uint64(value), uint64(iLogTable[decimal]))
		if hi != 0 {
			return append(b, "<si!OVERFLOW>"...)
		}
		whole = T(lo)
	}

	// By now we've bypassed the early exit cases,
	// what remains is to format our number.
	if isNegative {
		b = append(b, '-')
	}
	b = strconv.AppendInt(b, int64(whole), 10)
	if frac == 0 {
		return b
	}
	fraclen := ilog10(int64(frac))
	_ = fraclen
	b = strconv.AppendInt(b, 0, 10)
	return b
}

func abs[T fixed | ~int](a T) T {
	if a < 0 {
		return -a
	}
	return a
}
