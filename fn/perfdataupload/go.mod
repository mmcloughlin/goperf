module github.com/mmcloughlin/cb/fn/perfdataupload

go 1.13

replace github.com/mmcloughlin/cb => ../..

require (
	cloud.google.com/go/storage v1.6.0
	golang.org/x/build v0.0.0-20200304223525-ef9e68dfbdfe
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/perf v0.0.0-20200225203053-adf48cbc4550
	google.golang.org/api v0.18.0
)
