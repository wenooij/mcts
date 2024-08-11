package model

func PUCB(score, numRollouts, priorWeight, exploreTerm float64) float64 {
	nf := 1 / float64(numRollouts)
	return score*nf + priorWeight*exploreTerm*nf
}
