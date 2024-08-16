package model

func PUCB(score, numRollouts, priorWeight, exploreTerm float64) float64 {
	return (score + priorWeight*exploreTerm) / numRollouts
}
