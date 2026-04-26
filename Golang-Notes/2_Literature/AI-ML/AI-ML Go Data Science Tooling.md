---
tags:
  - type/note
  - theme/golang
  - theme/machine-learning
  - theme/deep-learning
aliases: []
lead: Complete curated list of Go packages for data science and ML — sourced directly from gopherdata/resources tooling README. Linear algebra, ML frameworks, NLP, visualization, time series, distributed pipelines, and what's still missing.
created: 2026-04-26
modified: 2026-04-27
source: "github.com/gopherdata/resources/tooling/README.md (Daniel Whitenack et al.); community/README.md."
---

# AI-ML — Go data science tooling

The gopherdata organization maintains a curated list of Go packages for data science. What follows is the full content of that README, organized and annotated. The author (Daniel Whitenack) is also the author of gophernet — see [[AI-ML Neural Network in Go]].

## Linear algebra and arithmetic

Everything in the ML implementation stack bottoms out here.

| Package | Purpose |
|---------|---------|
| `math` | Stdlib math functions |
| `math/cmplx` | Complex number constants and functions |
| `gonum.org/v1/gonum/floats` | `[]float64` helpers: `floats.Sum`, `floats.Max`, `floats.Dot`, norms |
| `gonum.org/v1/gonum/optimize` | Gradient descent, L-BFGS, Nelder-Mead and other optimizers |
| `go-hep.org/x/hep/fit` | WIP curve fitting and model fitting functions |

`gonum/floats` is used directly in gophernet — `floats.Sum` powers `sumAlongAxis` and `floats.Max` drives the argmax in accuracy evaluation.

## Matrices, arrays, and linear algebra

| Package | Purpose |
|---------|---------|
| `gonum.org/v1/gonum/mat` | Dense and sparse matrices — the core workhorse |
| `gonum.org/v1/gonum/blas` | BLAS routines (low-level backbone behind `mat`) |
| `gonum.org/v1/gonum/lapack` | Eigenvalues, SVD, linear solvers |
| `github.com/akualab/narray` | Multidimensional array package optimized with Go assembly |

`mat` is what [[AI-ML Neural Network in Go]] uses for every matrix operation. The LBDL's fully connected layers (see [[Model Components]] §4.2) map directly to `mat.Dense.Mul`.

## Data structures and munging

| Package | Purpose |
|---------|---------|
| `github.com/kniren/gota` | Dataframes — pandas-equivalent for Go |
| `github.com/Shixzie/ly` | Flexible DataFrame package aimed at ML workloads |
| `github.com/shuLhan/tabula` | Rows, columns, and matrix (table/dataset) manipulation |
| `github.com/gopherdata/gophernotes` | Go kernel for Jupyter notebooks |
| `github.com/kevinschoon/fit` | Toolkit for exploring and manipulating datasets |
| `neugram.io` | A programming language written in Go, designed for data munging |

Gota pays off during heavy pre-processing. For straightforward CSV-to-matrix pipelines like gophernet, `encoding/csv` from the stdlib is enough.

## CSV and tabular I/O

| Package | Purpose |
|---------|---------|
| `encoding/csv` | Stdlib CSV — sufficient for most ML data loading |
| `go-hep.org/x/hep/csvutil` | Convenient types and functions for CSV data files |
| `go-hep.org/x/hep/csvutil/csvdriver` | CSV as a `database/sql` driver |
| `github.com/dinedal/textql` | Run SQL queries directly against CSV or TSV files |
| `github.com/shuLhan/dsv` | Library for delimited separated value files |
| `github.com/frictionlessdata/tableschema-go` | Schema inference and table tooling for CSV |
| `github.com/j0hnsmith/csvplus` | CSV rows to/from typed structs with automatic type conversion |

## Other I/O formats

| Format | Package |
|--------|---------|
| JSON | `encoding/json` (stdlib), `tidwall/gjson` (fast queries), `pquerna/ffjson` (codegen) |
| NumPy `.npy` / `.npz` | `github.com/sbinet/npyio` |
| ARFF (Weka format) | `github.com/sbinet/go-arff` |
| Parquet | `github.com/xitongsys/parquet-go` |
| HDF5 | `gonum.org/v1/hdf5` (CGo bindings) |

`npyio` is worth knowing — pre-processed datasets frequently ship as `.npy` files and being able to load them directly saves a conversion step.

## General-purpose ML frameworks

| Package | Purpose |
|---------|---------|
| `github.com/chewxy/gorgonia` | Computation graph with autodiff — the autograd the LBDL describes in [[Training]] §3.4 |
| `github.com/sjwhitworth/golearn` | Batteries-included library with scikit-learn style API |
| `github.com/cdipaolo/goml` | ML algorithms plus online learning (incremental updates) |
| `github.com/xlvector/hector` | Binary classification algorithms |
| `github.com/shuLhan/go-mining` | CART, Random Forest, KNN, SMOTE resampling |
| `github.com/pa-m/sklearn` | WIP port of scikit-learn to Go |
| `github.com/galeone/tfgo` | TensorFlow + Go — run TF models from Go |
| `github.com/ctava/tfcgo` | Alternative TensorFlow C++ bridge |

Gorgonia is the serious option if you want autodiff without writing it yourself. It's verbose and the learning curve is steep — start with hand-coded gonum (see [[AI-ML Neural Network in Go]]) to understand what gorgonia is abstracting, then come back to it.

## Neural network specific

| Package | Purpose |
|---------|---------|
| `github.com/tleyden/neurgo` | Neural network toolkit |
| `github.com/fxsjy/gonn` | Backprop net (BPNN), radial basis function (RBF), probabilistic classifier (PCN) |
| `github.com/NOX73/go-neural` | Neural network in pure Go |
| `github.com/milosgajdos83/gosom` | Self-organizing maps — unsupervised competitive learning |
| `github.com/made2591/go-perceptron-go` | Single-layer perceptron; linear decision boundary only |

The `made2591/go-perceptron-go` package is the same author as the madeddu.xyz blog post referenced in the sources. The single-layer perceptron is the simplest possible case — no hidden layers, can only separate linearly separable problems.

## Classification

- `github.com/jbrukh/bayesian` — Naive Bayes
- `github.com/datastream/libsvm` — libsvm port for SVMs
- `github.com/barnjamin/randomforest` — Random Forest
- `github.com/rikonor/go-ann` — approximate k-nearest neighbor search

## Clustering

- `github.com/salkj/kmeans` — k-means
- `github.com/mpraski/clusters` — k-means++, DBSCAN, OPTICS

DBSCAN is useful when you don't know how many clusters to expect. k-means requires you to specify k up front.

## Regression

- `github.com/sajari/regression` — multivariable linear regression
- `github.com/glycerine/zettalm` — linear regression at very large scale

## Recommendation systems

- `github.com/jbochi/facts` — matrix factorization based recommendation
- `github.com/rlouf/birdland` — battle-tested recommendation library

## Probability, statistics, and experiments

| Package | Purpose |
|---------|---------|
| `gonum.org/v1/gonum/stat` | Full statistics package |
| `github.com/montanaflynn/stats` | Common functions missing from stdlib (median, percentiles, etc.) |
| `github.com/URXtech/planout-golang` | Multi-variate testing interpreter |
| `github.com/peleteiro/bandit-server` | Multi-armed bandit API for online learning |
| `github.com/dgryski/go-topk` | Streaming top-k via filtered space-saving |
| `github.com/dgryski/go-kll` | KLL sketch for streaming quantile approximation |
| `github.com/dgryski/go-linlog` | Linear-log bucketing and histograms |
| `github.com/dgryski/go-rbo` | Rank-biased overlap for comparing sorted result sets |

## NLP

| Package | Purpose |
|---------|---------|
| `github.com/jdkato/prose` | Most complete: tokenization, POS tagging, NER (pure Go) |
| `github.com/sajari/word2vec` | Query pre-trained word2vec embeddings |
| `github.com/ynqa/word-embedding` | Full word2vec and GloVe implementations |
| `github.com/cdipaolo/sentiment` | Sentiment analysis |
| `github.com/advancedlogic/go-freeling` | Partial port of Freeling 3.1 NLP toolkit |
| `github.com/endeveit/enca` | libenca bindings for encoding detection |
| `github.com/Lazin/go-ngram` | N-gram indexing |
| `github.com/reiver/go-porterstemmer` | Porter stemming algorithm |
| `github.com/kljensen/snowball` | Snowball stemmer |
| `github.com/blevesearch/segment` | Unicode text segmentation (Unicode Annex #29) |
| `github.com/jlubawy/go-gcnl` | Google Cloud Natural Language API client |
| `github.com/olebedev/when` | Natural language date/time parsing |
| `github.com/kampsy/gwizo` | Porter stemmer with additional features |
| `github.com/Shixzie/nlp` | General purpose NLP — parses text into filled model structs |
| `github.com/abadojack/whatlanggo` | Natural language detection |

## Time series

- `github.com/influxdata/influxdb` — open-source time series database
- `github.com/dgryski/go-holtwinters` — Holt-Winters forecasting (exponential smoothing)
- `github.com/dgryski/go-tsz` — time series compression from Facebook's Gorilla paper
- `github.com/dgryski/go-timewindow` — counters over sliding windows

## Graphs

- `gonum.org/v1/gonum/graph` — generalized graph package
- `github.com/gyuho/goraph` — graph data structures and algorithms
- `github.com/dgraph-io/dgraph` — distributed graph database
- `github.com/cayleygraph/cayley` — open-source graph database (inspired by Google Knowledge Graph)

## Bioinformatics

- `github.com/biogo` — bioinformatics library collection
- `github.com/ExaScience/elprep` — high-performance sequence alignment/map file preparation
- `github.com/MG-RAST/AWE` — workload management for bioinformatic workflow applications
- `github.com/kelvins/chronobiology` — biological temporal rhythm analysis

## Geospatial

- `github.com/golang/geo` — S2 geometry library in Go
- `github.com/twpayne/go-geom` — efficient geometry types, GeoJSON/KML/WKB encoding
- `github.com/twpayne/go-gpx` — read and write GPX documents
- `github.com/twpayne/go-kml` — create and write KML documents
- `github.com/twpayne/go-polyline` — Google Maps Polyline encoding and decoding

## Visualization and dashboards

- `gonum.org/v1/plot` — line, scatter, bar, histogram, CDF — outputs to PNG, SVG, PDF
- `github.com/ajstarks/svgo` — SVG generation
- `github.com/mmcloughlin/globe` — globe wireframe visualizations
- `github.com/gigablah/dashing-go` — real-time dashboards (port of Dashing)

Go's visualization tooling is behind Python's. For serious exploratory data analysis, gophernotes (Jupyter with a Go kernel) or piping data out to a Python plotting script is often more practical than generating plots directly in Go.

## Distributed data analysis and pipelines

| Package | Purpose |
|---------|---------|
| `github.com/pachyderm/pachyderm` | Versioned data pipelines with containerized transforms — git for data |
| `github.com/chrislusf/glow` | Distributed computation similar to Spark |
| `github.com/chrislusf/gleam` | Another distributed execution system |
| `github.com/flowbase/flowbase` | Flow-based programming micro-framework |
| `github.com/ExaScience/pargo` | Parallel algorithms using Go's primitives |
| `github.com/scipipe/scipipe` | Scientific workflow engine inspired by flow-based programming |
| `github.com/matryer/vice` | Go channels at horizontal scale via message queues |

`scipipe` is the most Go-idiomatic of these — pipelines are goroutines connected by typed channels. That maps directly onto [[Channels]] and [[Goroutines]] patterns. The LBDL covers why the compute pipeline model matters in [[Efficient Computation]].

## Databases

**SQL:** `lib/pq` (PostgreSQL), `jackc/pgx` (faster PostgreSQL + connection pooling), `go-pg/pg` (ORM), `go-sql-driver/mysql`, `mattn/go-sqlite3` (CGo), `lukasmartinelli/pgclimb` (PostgreSQL → JSON/CSV/XLSX export), `lukasmartinelli/pgfutter` (CSV/JSON → PostgreSQL import)

**NoSQL:** `gopkg.in/mgo.v2` (MongoDB), `gocql/gocql` (Cassandra), `go-redis/redis`, `garyburd/redigo` (Redis), `tsuna/gohbase` (HBase), `colinmarc/hdfs` (pure Go HDFS client)

## Web scraping

- `github.com/yhat/scrape` — higher-level web scraping interface
- `github.com/cathalgarvey/sqrape` — CSS selector scraping with Go reflection
- `github.com/PuerkitoBio/goquery` — jQuery-style HTML traversal
- `github.com/anaskhan96/soup` — BeautifulSoup-style interface
- `github.com/schollz/linkcrawler` — persistent distributed web crawler

## What's still missing (from the README's Proposed section)

The gopherdata community identified gaps that were never filled:

- **Multi-dimensional slices in the language** — no native `ndarray`. You're stuck with `[][]float64` or gonum matrices.
- **Robust concurrent minimization/fitting** — `gonum/optimize` exists but lacks histogram fitting and concurrent support.
- **Statistical modeling with nuisance parameters** — no Go equivalent of Stan or PyMC.
- **A/B testing package** — nothing native.
- **Datalog query system** for distributed computation — a Cascalog-equivalent that integrates with Go tools.

None of these exist as of 2026. The fundamental gap remains: there's no Go equivalent of PyTorch or JAX. No first-class GPU tensor library with autodiff built for training. Gorgonia is the closest thing, and it's a real undertaking.

Where Go makes sense for ML: inference servers and data pipelines, where you've trained in Python and need to serve fast with low memory overhead. ONNX runtime Go bindings (`onnxruntime-go`) let you run any ONNX-exported model from Go — that's the right production architecture in most cases.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[AI-ML Neural Network in Go]] — uses gonum/mat and gonum/floats directly
- see:: [[AI-ML Neural Network Foundations]] — the theory these tools implement
- see:: [[Go Packages and Modules]] — dependency management for third-party packages
- see:: [[Training]] — gorgonia implements autograd (§3.4); scaling laws context for why GPU matters (§3.7)
- see:: [[Model Components]] — the layer types that gorgonia and golearn implement
- see:: [[Channels]] — scipipe and flow-based pipelines are channel-based
- see:: [[Goroutines]] — distributed pipeline parallelism
- see:: [[Efficient Computation]] — hardware context for GPU/CUDA requirements

**Terms**
- gonum, mat.Dense, floats, gorgonia, autodiff, golearn, gota, gophernotes, prose, word2vec, npyio, HDF5, pachyderm, scipipe, ONNX, self-organizing map, k-means, DBSCAN, Holt-Winters, KLL sketch, S2 geometry, flow-based programming
