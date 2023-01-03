# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
GOPATH=$(shell go env GOPATH)

#Binary Stamping Vars
DATE=$(shell date +%Y.%m.%d//%T)
GITHASH=$(shell git rev-parse HEAD)
GITBRANCH=$(shell git branch --show-current)
GITHASHDATE=$(shell git show -s --format=%ci HEAD | sed 's/ /\//g')

#File building dependencies
FILEDEPS = main.go ast.go semantic.go repl.go ocli.go aststr.go \
 astnum.go astbool.go astflow.go astutil.go completer.go parser.go

main: $(FILEDEPS)
	go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	$(FILEDEPS)

#OTHER PLATFORM COMPILATION BLOCK
mac: $(FILEDEPS)
	GOOS=darwin go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	$(FILEDEPS)
	
win: $(FILEDEPS)
	GOOS=windows go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	$(FILEDEPS)
	
clean:
	rm main

