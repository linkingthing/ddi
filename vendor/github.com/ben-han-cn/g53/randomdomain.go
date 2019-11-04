package g53

import (
	"bytes"
	"math/rand"

	"github.com/ben-han-cn/cement/randomdata"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789-"

//FQDN  www.baidu.com. strlen = wirelen - 1
//NON-FQDN  www.baidu.com strlen = wirelen - 2
func RandomNoneFQDNDomain() string {
	domainLen := 1 + rand.Intn(MAX_WIRE-3)
	labelCnt := 1 + rand.Intn(MAX_LABELS-2)
	generatedLen := 0
	var buf bytes.Buffer
	buf.Grow(domainLen)

	for i := 0; i < labelCnt; i++ {
		maxLabelLen := MAX_LABEL_LEN - 1
		leftLen := domainLen - generatedLen
		if maxLabelLen > leftLen {
			maxLabelLen = leftLen
		}

		isLastLabel := (i + 1) == labelCnt
		var labelLen int
		if maxLabelLen == 1 {
			labelLen = 1
		} else if isLastLabel {
			labelLen = maxLabelLen
		} else {
			labelLen = 1 + rand.Intn(maxLabelLen-1)
		}

		buf.WriteString(randomdata.RandStringWithLetter(labelLen, letterBytes))
		generatedLen = generatedLen + labelLen

		if generatedLen < domainLen {
			if generatedLen+1 == domainLen {
				buf.WriteString(randomdata.RandStringWithLetter(1, letterBytes))
			} else if isLastLabel == false {
				buf.WriteString(".")
			}
			generatedLen += 1
		}

		if generatedLen == domainLen {
			break
		}
	}

	return buf.String()
}
