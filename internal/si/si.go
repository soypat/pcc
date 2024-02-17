package si

// NewDimension creates a new dimension from the given exponents.
func NewDimension(len, mass, time, temp, icurrent, luminosity, amount, _ int8) Dimension {
	return Dimension{
		dims: [4]byte{
			i8ToI4(len) | i8ToI4(mass)<<4,
			i8ToI4(time) | i8ToI4(temp)<<4,
			i8ToI4(icurrent) | i8ToI4(luminosity)<<4,
			i8ToI4(amount) | i8ToI4(0)<<4,
		},
	}
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

func (d Dimension) ExpLength() int8      { return i4ToI8(d.dims[0] & 0xf) }
func (d Dimension) ExpMass() int8        { return i4ToI8(d.dims[0] >> 4) }
func (d Dimension) ExpTime() int8        { return i4ToI8(d.dims[1] & 0xf) }
func (d Dimension) ExpTemperature() int8 { return i4ToI8(d.dims[1] >> 4) }
func (d Dimension) ExpCurrent() int8     { return i4ToI8(d.dims[2] & 0xf) }
func (d Dimension) ExpLuminous() int8    { return i4ToI8(d.dims[2] >> 4) }
func (d Dimension) ExpAmount() int8      { return i4ToI8(d.dims[3] & 0xf) }

func i8ToI4(c int8) byte {
	if c < 0 {
		c |= 1 << 3
	}
	return byte(c) & 0xf
}

// i4ToI8 converts lower 4 bits of a byte to a signed 4 bit integer.
func i4ToI8(c byte) int8 {
	c &= 0xf
	if c&0x8 != 0 {
		c = c | 0xf0
	}
	return int8(c)
}
