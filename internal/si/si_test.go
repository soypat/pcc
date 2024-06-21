package si

import "testing"

func TestNewDimension(t *testing.T) {
	for l := -maxunit; l < maxunit; l += 11 {
		for m := -maxunit; m < maxunit; m += 11 {
			for k := -maxunit; k < maxunit; k += 11 {
				d, err := NewDimension(l, m, k, 4, 5, 6, 7)
				if err != nil {
					t.Fatal("dimension error", err)
				}
				gotl := d.ExpLength()
				if gotl != l {
					t.Errorf("%s:L want %d, got %d", d.String(), l, gotl)
				}
				gotm := d.ExpMass()
				if gotm != m {
					t.Errorf("%s:M want %d, got %d", d.String(), m, gotm)
				}
				gotk := d.ExpTime()
				if gotk != k {
					t.Errorf("%s:K want %d, got %d", d.String(), k, gotk)
				}
				if t.Failed() {
					t.Fatal(t.Name(), "exit early")
				}
			}
		}
	}
}

func TestA(t *testing.T) {
	d, err := NewDimension(1, 2, 3, 4, 5, 6, 6)
	if err != nil {
		panic(err)
	}
	if d.String() != "LM²T³K⁴I⁵J⁶N⁶" {
		t.Fatal("unexpected string positives", d.String())
	}

	d, err = NewDimension(-1, -2, -3, -4, -5, -6, -6)
	if err != nil {
		panic(err)
	}
	if d.String() != "L⁻¹M⁻²T⁻³K⁻⁴I⁻⁵J⁻⁶N⁻⁶" {
		t.Fatal("unexpected string negatives", d.String())
	}
}
