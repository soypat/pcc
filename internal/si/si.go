package si

import (
	"errors"
	"math/bits"
	"strconv"
)

const (
	maxunit = 7
)

// NewDimension creates a new dimension from the given exponents.
func NewDimension(length, mass, time, temperature, electricCurrent, luminosity, amount, UNUSED int) (Dimension, error) {
	if isDimOOB(length) || isDimOOB(mass) || isDimOOB(time) ||
		isDimOOB(temperature) || isDimOOB(electricCurrent) ||
		isDimOOB(luminosity) || isDimOOB(amount) {
		return Dimension{}, errors.New("overflow dimension storage")
	} else if UNUSED != 0 {
		return Dimension{}, errors.New("use of reserved dimension")
	}
	return Dimension{
		dims: [4]byte{
			intToI4(length) | intToI4(mass)<<4,
			intToI4(time) | intToI4(temperature)<<4,
			intToI4(electricCurrent) | intToI4(luminosity)<<4,
			intToI4(amount) | intToI4(0)<<4,
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
	dims [4]byte
}

func (d Dimension) ExpLength() int      { return i4ToInt(d.dims[0] & 0xf) }
func (d Dimension) ExpMass() int        { return i4ToInt(d.dims[0] >> 4) }
func (d Dimension) ExpTime() int        { return i4ToInt(d.dims[1] & 0xf) }
func (d Dimension) ExpTemperature() int { return i4ToInt(d.dims[1] >> 4) }
func (d Dimension) ExpCurrent() int     { return i4ToInt(d.dims[2] & 0xf) }
func (d Dimension) ExpLuminous() int    { return i4ToInt(d.dims[2] >> 4) }
func (d Dimension) ExpAmount() int      { return i4ToInt(d.dims[3] & 0xf) }

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
	l := a.ExpLength() + b.ExpLength()
	m := a.ExpMass() + b.ExpMass()
	t := a.ExpTime() + b.ExpTime()
	T := a.ExpTemperature() + b.ExpTemperature()
	i := a.ExpCurrent() + b.ExpCurrent()
	L := a.ExpLuminous() + b.ExpLuminous()
	q := a.ExpAmount() + b.ExpAmount()
	return NewDimension(l, m, t, T, i, L, q, 0)
}

func DivDim(a, b Dimension) (Dimension, error) {
	return MulDim(a, b.Inv())
}

func intToI4(c int) byte {
	if c < 0 {
		c |= 1 << 3
	}
	return byte(c) & 0xf
}

// i4ToInt converts lower 4 bits of a byte to a signed 4 bit integer.
func i4ToInt(c byte) int {
	c &= 0xf
	if c&0x8 != 0 {
		c = c | 0xf0
	}
	return int(c)
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
func (p Prefix) Character() (s rune) {
	if p == PrefixMicro {
		s = 'μ'
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
