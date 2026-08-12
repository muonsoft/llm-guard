# Benchmark baseline (M7 release candidate)

Representative offline benchmarks for Detect, Mask, and Restore. Numbers are a
reproducible development baseline only — **not** a production SLO or performance guarantee.

## Environment

| Field | Value |
|-------|-------|
| Date | 2026-08-12 |
| Go | go1.26.2 linux/amd64 |
| OS | Linux 5.15.0-181-generic (Ubuntu) |
| CPU | AMD EPYC-Rome Processor |
| Package | `github.com/muonsoft/llm-guard` |

## Command

```bash
go test ./... -run '^$' -bench . -benchmem -count=5
```

Benchmark inputs (see `benchmark_test.go`):

| Name | Operation | Input |
|------|-----------|-------|
| `RUPrompt` | Detect/Mask/Restore | Russian PERSON + ADDRESS sentence |
| `MixedPII` | Detect/Mask/Restore | RU phone, email, INN mix |
| `SyntheticSecret` | Detect/Mask/Restore | Synthetic JWT-like token |

Mask/Restore setup (TokenSet, masked text) is performed outside timed loops.

## Raw results (`-count=5`)

```
goos: linux
goarch: amd64
pkg: github.com/muonsoft/llm-guard
cpu: AMD EPYC-Rome Processor
BenchmarkDetect_RUPrompt-4           	    5221	    254572 ns/op	   11850 B/op	     100 allocs/op
BenchmarkDetect_RUPrompt-4           	    6546	    172918 ns/op	   11876 B/op	     100 allocs/op
BenchmarkDetect_RUPrompt-4           	    6268	    173766 ns/op	   11900 B/op	     100 allocs/op
BenchmarkDetect_RUPrompt-4           	    9847	    150266 ns/op	   11856 B/op	     100 allocs/op
BenchmarkDetect_RUPrompt-4           	    7456	    222449 ns/op	   11879 B/op	     100 allocs/op
BenchmarkDetect_MixedPII-4           	    6873	    220320 ns/op	   15956 B/op	     106 allocs/op
BenchmarkDetect_MixedPII-4           	    6913	    183564 ns/op	   15965 B/op	     106 allocs/op
BenchmarkDetect_MixedPII-4           	    7908	    216495 ns/op	   16007 B/op	     106 allocs/op
BenchmarkDetect_MixedPII-4           	    5980	    233010 ns/op	   15906 B/op	     106 allocs/op
BenchmarkDetect_MixedPII-4           	    5996	    229830 ns/op	   15875 B/op	     106 allocs/op
BenchmarkDetect_SyntheticSecret-4    	    6013	    177633 ns/op	   11145 B/op	     102 allocs/op
BenchmarkDetect_SyntheticSecret-4    	    5858	    190074 ns/op	   11211 B/op	     102 allocs/op
BenchmarkDetect_SyntheticSecret-4    	    5751	    217638 ns/op	   11137 B/op	     102 allocs/op
BenchmarkDetect_SyntheticSecret-4    	    5558	    186127 ns/op	   11214 B/op	     102 allocs/op
BenchmarkDetect_SyntheticSecret-4    	    7621	    176056 ns/op	   11170 B/op	     102 allocs/op
BenchmarkMask_RUPrompt-4             	    5730	    210344 ns/op	   13579 B/op	     131 allocs/op
BenchmarkMask_RUPrompt-4             	    6717	    219353 ns/op	   13621 B/op	     131 allocs/op
BenchmarkMask_RUPrompt-4             	    5758	    238869 ns/op	   13643 B/op	     131 allocs/op
BenchmarkMask_RUPrompt-4             	    4923	    207923 ns/op	   13608 B/op	     131 allocs/op
BenchmarkMask_RUPrompt-4             	    5156	    233774 ns/op	   13616 B/op	     131 allocs/op
BenchmarkMask_MixedPII-4             	    4422	    226739 ns/op	   18320 B/op	     139 allocs/op
BenchmarkMask_MixedPII-4             	    5281	    241090 ns/op	   18315 B/op	     139 allocs/op
BenchmarkMask_MixedPII-4             	    6208	    183215 ns/op	   18276 B/op	     139 allocs/op
BenchmarkMask_MixedPII-4             	    9680	    223599 ns/op	   18325 B/op	     139 allocs/op
BenchmarkMask_MixedPII-4             	    5860	    215108 ns/op	   18238 B/op	     139 allocs/op
BenchmarkMask_SyntheticSecret-4      	    5721	    282804 ns/op	   11965 B/op	     124 allocs/op
BenchmarkMask_SyntheticSecret-4      	    4814	    208569 ns/op	   11998 B/op	     124 allocs/op
BenchmarkMask_SyntheticSecret-4      	    5648	    202163 ns/op	   11986 B/op	     124 allocs/op
BenchmarkMask_SyntheticSecret-4      	    5700	    203296 ns/op	   11958 B/op	     124 allocs/op
BenchmarkMask_SyntheticSecret-4      	    4862	    222424 ns/op	   11989 B/op	     124 allocs/op
BenchmarkRestore_MixedPII-4          	  156102	      7583 ns/op	    1832 B/op	      17 allocs/op
BenchmarkRestore_MixedPII-4          	  185757	      6842 ns/op	    1832 B/op	      17 allocs/op
BenchmarkRestore_MixedPII-4          	  268766	      5098 ns/op	    1832 B/op	      17 allocs/op
BenchmarkRestore_MixedPII-4          	  185278	      7325 ns/op	    1832 B/op	      17 allocs/op
BenchmarkRestore_MixedPII-4          	  161305	      6528 ns/op	    1832 B/op	      17 allocs/op
BenchmarkRestore_RUPrompt-4          	  212836	      6203 ns/op	    1672 B/op	      15 allocs/op
BenchmarkRestore_RUPrompt-4          	  168232	      6233 ns/op	    1672 B/op	      15 allocs/op
BenchmarkRestore_RUPrompt-4          	  216666	      6265 ns/op	    1640 B/op	      15 allocs/op
BenchmarkRestore_RUPrompt-4          	  224161	      4808 ns/op	    1640 B/op	      15 allocs/op
BenchmarkRestore_RUPrompt-4          	  226780	      6041 ns/op	    1672 B/op	      15 allocs/op
BenchmarkRestore_SyntheticSecret-4   	  213873	      5018 ns/op	    2872 B/op	       6 allocs/op
BenchmarkRestore_SyntheticSecret-4   	  213585	      5826 ns/op	    2872 B/op	       6 allocs/op
BenchmarkRestore_SyntheticSecret-4   	  244614	      5608 ns/op	    2872 B/op	       6 allocs/op
BenchmarkRestore_SyntheticSecret-4   	  273679	      4352 ns/op	    2872 B/op	       6 allocs/op
BenchmarkRestore_SyntheticSecret-4   	  398019	      4041 ns/op	    2872 B/op	       6 allocs/op
BenchmarkObserver_DefaultNoop-4       	  126669	     12633 ns/op	     994 B/op	      18 allocs/op
BenchmarkObserver_DefaultNoop-4       	  112971	     16328 ns/op	     996 B/op	      18 allocs/op
BenchmarkObserver_DefaultNoop-4       	   72945	     16768 ns/op	     991 B/op	      18 allocs/op
BenchmarkObserver_DefaultNoop-4       	  129471	     12968 ns/op	     993 B/op	      18 allocs/op
BenchmarkObserver_DefaultNoop-4       	  104143	     15699 ns/op	     995 B/op	      18 allocs/op
BenchmarkObserver_WithCallback-4      	   86775	     14263 ns/op	     995 B/op	      18 allocs/op
BenchmarkObserver_WithCallback-4      	   75224	     13845 ns/op	     997 B/op	      18 allocs/op
BenchmarkObserver_WithCallback-4      	   97953	     14049 ns/op	     997 B/op	      18 allocs/op
BenchmarkObserver_WithCallback-4      	  104005	     16229 ns/op	     993 B/op	      18 allocs/op
BenchmarkObserver_WithCallback-4      	   94646	     14104 ns/op	     993 B/op	      18 allocs/op
```

## Caveats

- Results vary by CPU, load, and Go version; compare by stable benchmark names, not absolute ns/op across machines.
- No network, exporter, or external services are involved.
- Observer overhead benchmark uses an in-memory callback; production integrations may differ.
