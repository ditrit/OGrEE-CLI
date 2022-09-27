# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
GOPATH=$(shell go env GOPATH)
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex
DATE=$(shell date +%Y.%m.%d//%T)
GITHASH=$(shell git rev-parse HEAD)
GITBRANCH=$(shell git branch --show-current)
GITHASHDATE=$(shell git show -s --format=%ci HEAD | sed 's/ /\//g')


main: interpreter main.go ast.go lexer.nn.go y.go repl.go aststr.go astnum.go astbool.go astflow.go astutil.go completer.go
	go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	main.go ast.go lexer.nn.go y.go repl.go aststr.go astnum.go astbool.go astflow.go astutil.go completer.go
	

interpreter: parser lexer buildTimeScript

parser: interpreter/parser.y controllers/commandController.go
	$(GOYACC) "interpreter/parser.y" 

lexer: interpreter/lexer.nex
	$(NEX) "interpreter/lexer.nex"; mv interpreter/lexer.nn.go .

buildTimeScript:
	$(info Injecting build time code...)
	other/injectionscript.py

#OTHER PLATFORM COMPILATION BLOCK
mac: interpreter main.go ast.go lexer.nn.go y.go repl.go aststr.go astnum.go astbool.go astflow.go astutil.go completer.go
	GOOS=darwin go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	main.go ast.go lexer.nn.go y.go repl.go aststr.go astnum.go astbool.go astflow.go astutil.go completer.go
	

win: interpreter main.go ast.go lexer.nn.go y.go repl.go aststr.go astnum.go astbool.go astflow.go astutil.go completer.go
	GOOS=windows go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	main.go ast.go lexer.nn.go y.go repl.go aststr.go astnum.go astbool.go astflow.go astutil.go completer.go
	

clean:
	rm main y.go lexer.nn.go y.output parser.tab.c

