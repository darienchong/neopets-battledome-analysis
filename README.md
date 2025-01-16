# What
A project written in Go that does analysis of datamined and empirical Neopets Battledome drop data. It can do
- Simulation of Battledome item drops based on datamined item weights from [here](https://www.reddit.com/r/neopets/comments/yc488a/battledome_dome_loot/)
- Analysis of Battledome drop data (mean profit, confidence intervals using the [Clopper-Pearson interval method](https://en.wikipedia.org/wiki/Binomial_proportion_confidence_interval))
- Comparison of the above two analyses
- Comparison of mean profit across Battledome challengers, arenas
- Profit breakdown (profit contribution, expected profit per item, etc.) on an arena/challenger level
- And more...? 

# Example of program output
![Example program output](https://github.com/darienchong/neopets-battledome-analysis/blob/master/example.png?raw=true)
![Example program output 2](https://github.com/darienchong/neopets-battledome-analysis/blob/master/example2.png?raw=true)

# Roadmap
- [x] [Add arena comparison](https://github.com/darienchong/neopets-battledome-analysis/commit/146edd8d8014ab56d39e4fbb014bfd698d73df3a)
- [x] [Add challenger comparison](https://github.com/darienchong/neopets-battledome-analysis/commit/724c4c6986900cdaa751a98b1ff00d31f74d3b42)
- [x] [Add single arena profit breakdown](https://github.com/darienchong/neopets-battledome-analysis/commit/4f1a7b98236b3455ebb901e1655ca3f99ba24cb4)
- [x] [Add single challenger profit breakdown](https://github.com/darienchong/neopets-battledome-analysis/commit/4f1a7b98236b3455ebb901e1655ca3f99ba24cb4)
- [x] [Add stack traces in error propagation](https://github.com/darienchong/neopets-battledome-analysis/commit/0ebf6b0b4d195a46be78b7052e8e28619268d295)
- [x] [Filter out challenger-specific drops in arena comparison](https://github.com/darienchong/neopets-battledome-analysis/commit/4931277352ca8c0ca04d50dbd0a94037144bdc72)
- [ ] What's next?