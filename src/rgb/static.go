package rgb

// Static will run RGB function
func (r *ActiveRGB) Static() {
	buf := map[int][]byte{}
	modify := ModifyBrightness(*r.RGBStartColor)
	for j := 0; j < r.LightChannels; j++ {
		buf[j] = []byte{
			byte(modify.Red),
			byte(modify.Green),
			byte(modify.Blue),
		}
		if r.IsAIO && r.HasLCD {
			if j > 15 && j < 20 {
				buf[j] = []byte{0, 0, 0}
			}
		}
	}
	if r.Inverted {
		r.Output = SetColorInverted(buf)
	} else {
		r.Output = SetColor(buf)
	}
}
