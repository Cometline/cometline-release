package memory

import (
	"context"
	"sort"
	"strings"
	"time"
)

type retriever struct {
	store    *store
	embedder Embedder
	settings Settings
}

func (r *retriever) retrieve(ctx context.Context, query string, maxN int, threshold float64) ([]ScoredMemory, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	memories, err := r.store.listActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(memories) == 0 {
		return nil, nil
	}

	recordByID := make(map[string]Record, len(memories))
	for _, m := range memories {
		recordByID[m.ID] = m
	}

	simByID := make(map[string]float64)
	vecs, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(vecs) > 0 {
		qvec := vecs[0]
		for _, m := range memories {
			if len(m.Embedding) == 0 {
				continue
			}
			simByID[m.ID] = cosineSimilarity(qvec, m.Embedding)
		}
	}

	vectorRanked := topNBySimilarity(simByID, retrievalPoolSize)

	ftsRanked, err := r.store.searchFTS(ctx, query, retrievalPoolSize)
	if err != nil {
		return nil, err
	}
	ftsHit := make(map[string]struct{}, len(ftsRanked))
	for _, id := range ftsRanked {
		ftsHit[id] = struct{}{}
	}

	rrfScores := reciprocalRankFusion(vectorRanked, ftsRanked)
	if len(rrfScores) == 0 {
		return nil, nil
	}

	now := time.Now()
	var scored []ScoredMemory
	for id, rrfScore := range rrfScores {
		m, ok := recordByID[id]
		if !ok {
			continue
		}
		sim := simByID[id]
		if sim < threshold {
			if _, fts := ftsHit[id]; !fts {
				continue
			}
		}
		ew := EffectiveWeight(m, now, r.settings.Lifecycle)
		scored = append(scored, ScoredMemory{
			Record:          m,
			Similarity:      sim,
			EffectiveWeight: ew,
			RetrievalScore:  rrfScore * ew,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].RetrievalScore > scored[j].RetrievalScore
	})
	if maxN <= 0 {
		maxN = r.settings.MaxRetrieved
	}
	if maxN > 0 && len(scored) > maxN {
		scored = scored[:maxN]
	}

	for _, sm := range scored {
		_ = r.store.touchAccess(ctx, sm.ID)
		_ = r.store.logEvent(ctx, sm.ID, "inject", "")
	}
	return scored, nil
}

func (r *retriever) search(ctx context.Context, query string, maxN int) ([]ScoredMemory, error) {
	return r.retrieve(ctx, query, maxN, 0)
}

func (r *retriever) bestMatch(ctx context.Context, vec []float32) (Record, float64, error) {
	memories, err := r.store.listActive(ctx)
	if err != nil {
		return Record{}, 0, err
	}
	var best Record
	var bestSim float64
	for _, m := range memories {
		if len(m.Embedding) == 0 {
			continue
		}
		sim := cosineSimilarity(vec, m.Embedding)
		if sim > bestSim {
			bestSim = sim
			best = m
		}
	}
	return best, bestSim, nil
}
