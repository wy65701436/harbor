# Makefile for Harbor project
#
# Targets:
#
# all:			prepare env, compile binaries, build images and install images
# prepare: 		prepare env
# compile: 		compile adminserver, ui and jobservice code
#
# compile_golangimage:
#			compile from golang image
#			for example: make compile_golangimage -e GOBUILDIMAGE= \
#							golang:1.11.2
# compile_adminserver, compile_core, compile_jobservice: compile specific binary
#
# build:	build Harbor docker images from photon baseimage
#
# install:		include compile binarys, build images, prepare specific \
#				version composefile and startup Harbor instance
#
# start:		startup Harbor instance
#
# down:			shutdown Harbor instance
#
# package_online:
#				prepare online install package
#			for example: make package_online -e DEVFLAG=false\
#							REGISTRYSERVER=reg-bj.goharbor.io \
#							REGISTRYPROJECTNAME=harborrelease
#
# package_offline:
#				prepare offline install package
#
# pushimage:	push Harbor images to specific registry server
#			for example: make pushimage -e DEVFLAG=false REGISTRYUSER=admin \
#							REGISTRYPASSWORD=***** \
#							REGISTRYSERVER=reg-bj.goharbor.io/ \
#							REGISTRYPROJECTNAME=harborrelease
#				note**: need add "/" on end of REGISTRYSERVER. If not setting \
#						this value will push images directly to dockerhub.
#						 make pushimage -e DEVFLAG=false REGISTRYUSER=goharbor \
#							REGISTRYPASSWORD=***** \
#							REGISTRYPROJECTNAME=goharbor
#
# clean:        remove binary, Harbor images, specific version docker-compose \
#               file, specific version tag and online/offline install package
# cleanbinary:	remove adminserver, ui and jobservice binary
# cleanimage: 	remove Harbor images
# cleandockercomposefile:
#				remove specific version docker-compose
# cleanversiontag:
#				cleanpackageremove specific version tag
# cleanpackage: remove online/offline install package
#
# other example:
#	clean specific version binarys and images:
#				make clean -e VERSIONTAG=[TAG]
#				note**: If commit new code to github, the git commit TAG will \
#				change. Better use this commond clean previous images and \
#				files with specific TAG.
#   By default DEVFLAG=true, if you want to release new version of Harbor, \
#		should setting the flag to false.
#				make XXXX -e DEVFLAG=false

SHELL := /bin/bash
BUILDPATH=$(CURDIR)
MAKEPATH=$(BUILDPATH)/make
MAKEDEVPATH=$(MAKEPATH)/dev
SRCPATH=./src
TOOLSPATH=$(BUILDPATH)/tools
CORE_PATH=$(BUILDPATH)/src/core
PORTAL_PATH=$(BUILDPATH)/src/portal
GOBASEPATH=/go/src/github.com/goharbor
CHECKENVCMD=checkenv.sh

# parameters
REGISTRYSERVER=
REGISTRYPROJECTNAME=goharbor
DEVFLAG=true
NOTARYFLAG=false
CLAIRFLAG=false
HTTPPROXY=
BUILDBIN=false
MIGRATORFLAG=false
# enable/disable chart repo supporting
CHARTFLAG=false

# version prepare
# for docker image tag
VERSIONTAG=dev
# for harbor package name
PKGVERSIONTAG=dev
# for harbor about dialog
UIVERSIONTAG=dev
VERSIONFILEPATH=$(CURDIR)
VERSIONFILENAME=UIVERSION

#versions
REGISTRYVERSION=v2.6.2
NGINXVERSION=$(VERSIONTAG)
NOTARYVERSION=v0.6.1
CLAIRVERSION=v2.0.7
CLAIRDBVERSION=$(VERSIONTAG)
MIGRATORVERSION=$(VERSIONTAG)
REDISVERSION=$(VERSIONTAG)

# version of chartmuseum
CHARTMUSEUMVERSION=v0.7.1

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMIMAGE=$(DOCKERCMD) rmi
DOCKERPULL=$(DOCKERCMD) pull
DOCKERIMASES=$(DOCKERCMD) images
DOCKERSAVE=$(DOCKERCMD) save
DOCKERCOMPOSECMD=$(shell which docker-compose)
DOCKERTAG=$(DOCKERCMD) tag

# go parameters
GOCMD=$(shell which go)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
GOBUILDIMAGE=golang:1.11.2
GOBUILDPATH=$(GOBASEPATH)/harbor
GOIMAGEBUILDCMD=/usr/local/go/bin/go
GOIMAGEBUILD=$(GOIMAGEBUILDCMD) build
GOBUILDPATH_ADMINSERVER=$(GOBUILDPATH)/src/adminserver
GOBUILDPATH_CORE=$(GOBUILDPATH)/src/core
GOBUILDPATH_JOBSERVICE=$(GOBUILDPATH)/src/jobservice
GOBUILDPATH_REGISTRYCTL=$(GOBUILDPATH)/src/registryctl
GOBUILDMAKEPATH=$(GOBUILDPATH)/make
GOBUILDMAKEPATH_ADMINSERVER=$(GOBUILDMAKEPATH)/photon/adminserver
GOBUILDMAKEPATH_CORE=$(GOBUILDMAKEPATH)/photon/core
GOBUILDMAKEPATH_JOBSERVICE=$(GOBUILDMAKEPATH)/photon/jobservice
GOBUILDMAKEPATH_REGISTRYCTL=$(GOBUILDMAKEPATH)/photon/registryctl

# binary
ADMINSERVERBINARYPATH=$(MAKEDEVPATH)/adminserver
ADMINSERVERBINARYNAME=harbor_adminserver
CORE_BINARYPATH=$(MAKEDEVPATH)/core
CORE_BINARYNAME=harbor_core
JOBSERVICEBINARYPATH=$(MAKEDEVPATH)/jobservice
JOBSERVICEBINARYNAME=harbor_jobservice
REGISTRYCTLBINARYPATH=$(MAKEDEVPATH)/registryctl
REGISTRYCTLBINARYNAME=harbor_registryctl

# configfile
CONFIGPATH=$(MAKEPATH)
CONFIGFILE=harbor.cfg

# prepare parameters
PREPAREPATH=$(TOOLSPATH)
PREPARECMD=prepare
PREPARECMD_PARA=--conf $(CONFIGPATH)/$(CONFIGFILE)
ifeq ($(NOTARYFLAG), true)
	PREPARECMD_PARA+= --with-notary
endif
ifeq ($(CLAIRFLAG), true)
	PREPARECMD_PARA+= --with-clair
endif
# append chartmuseum parameters if set
ifeq ($(CHARTFLAG), true)
    PREPARECMD_PARA+= --with-chartmuseum
endif

# makefile
MAKEFILEPATH_PHOTON=$(MAKEPATH)/photon

# common dockerfile
DOCKERFILEPATH_COMMON=$(MAKEPATH)/common

# docker image name
DOCKERIMAGENAME_ADMINSERVER=goharbor/harbor-adminserver
DOCKERIMAGENAME_PORTAL=goharbor/harbor-portal
DOCKERIMAGENAME_CORE=goharbor/harbor-core
DOCKERIMAGENAME_JOBSERVICE=goharbor/harbor-jobservice
DOCKERIMAGENAME_LOG=goharbor/harbor-log
DOCKERIMAGENAME_DB=goharbor/harbor-db
DOCKERIMAGENAME_CHART_SERVER=goharbor/chartmuseum-photon
DOCKERIMAGENAME_REGCTL=goharbor/harbor-registryctl

# docker-compose files
DOCKERCOMPOSEFILEPATH=$(MAKEPATH)
DOCKERCOMPOSETPLFILENAME=docker-compose.tpl
DOCKERCOMPOSEFILENAME=docker-compose.yml
DOCKERCOMPOSENOTARYTPLFILENAME=docker-compose.notary.tpl
DOCKERCOMPOSENOTARYFILENAME=docker-compose.notary.yml
DOCKERCOMPOSECLAIRTPLFILENAME=docker-compose.clair.tpl
DOCKERCOMPOSECLAIRFILENAME=docker-compose.clair.yml
DOCKERCOMPOSECHARTMUSEUMTPLFILENAME=docker-compose.chartmuseum.tpl
DOCKERCOMPOSECHARTMUSEUMFILENAME=docker-compose.chartmuseum.yml

SEDCMD=$(shell which sed)

# package
TARCMD=$(shell which tar)
ZIPCMD=$(shell which gzip)
DOCKERIMGFILE=harbor
HARBORPKG=harbor

# pushimage
PUSHSCRIPTPATH=$(MAKEPATH)
PUSHSCRIPTNAME=pushimage.sh
REGISTRYUSER=user
REGISTRYPASSWORD=default

# cmds
DOCKERSAVE_PARA=$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(DOCKERIMAGENAME_REGCTL):$(VERSIONTAG) \
		goharbor/redis-photon:$(REDISVERSION) \
		goharbor/nginx-photon:$(NGINXVERSION) goharbor/registry-photon:$(REGISTRYVERSION)-$(VERSIONTAG)

PACKAGE_OFFLINE_PARA=-zcvf harbor-offline-installer-$(PKGVERSIONTAG).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar.gz \
				  $(HARBORPKG)/prepare \
				  $(HARBORPKG)/LICENSE $(HARBORPKG)/install.sh \
				  $(HARBORPKG)/harbor.cfg $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/open_source_license 

PACKAGE_ONLINE_PARA=-zcvf harbor-online-installer-$(PKGVERSIONTAG).tgz \
		          $(HARBORPKG)/common/templates $(HARBORPKG)/prepare \
				  $(HARBORPKG)/LICENSE \
				  $(HARBORPKG)/install.sh $(HARBORPKG)/$(DOCKERCOMPOSEFILENAME) \
				  $(HARBORPKG)/harbor.cfg \
				  $(HARBORPKG)/open_source_license
				  
DOCKERCOMPOSE_LIST=-f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

ifeq ($(NOTARYFLAG), true)
	DOCKERSAVE_PARA+= goharbor/notary-server-photon:$(NOTARYVERSION)-$(VERSIONTAG) goharbor/notary-signer-photon:$(NOTARYVERSION)-$(VERSIONTAG)
	PACKAGE_OFFLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSENOTARYFILENAME)
	PACKAGE_ONLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSENOTARYFILENAME)
	DOCKERCOMPOSE_LIST+= -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)
endif
ifeq ($(CLAIRFLAG), true)
	DOCKERSAVE_PARA+= goharbor/clair-photon:$(CLAIRVERSION)-$(VERSIONTAG)
	PACKAGE_OFFLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSECLAIRFILENAME)
	PACKAGE_ONLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSECLAIRFILENAME)
	DOCKERCOMPOSE_LIST+= -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
endif
ifeq ($(MIGRATORFLAG), true)
	DOCKERSAVE_PARA+= goharbor/harbor-migrator:$(MIGRATORVERSION)
endif
# append chartmuseum parameters if set
ifeq ($(CHARTFLAG), true)
	DOCKERSAVE_PARA+= $(DOCKERIMAGENAME_CHART_SERVER):$(CHARTMUSEUMVERSION)-$(VERSIONTAG)
	PACKAGE_OFFLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSECHARTMUSEUMFILENAME)
	PACKAGE_ONLINE_PARA+= $(HARBORPKG)/$(DOCKERCOMPOSECHARTMUSEUMFILENAME)
	DOCKERCOMPOSE_LIST+= -f $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECHARTMUSEUMFILENAME)
endif

ui_version:
	@printf $(UIVERSIONTAG) > $(VERSIONFILEPATH)/$(VERSIONFILENAME);

check_environment:
	@$(MAKEPATH)/$(CHECKENVCMD)

compile_adminserver:
	@echo "compiling binary for adminserver (golang image)..."
	@echo $(GOBASEPATH)
	@echo $(GOBUILDPATH)
	$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_ADMINSERVER) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_ADMINSERVER)/$(ADMINSERVERBINARYNAME)
	@echo "Done."

compile_core:
	@echo "compiling binary for core (golang image)..."
	@echo $(GOBASEPATH)
	@echo $(GOBUILDPATH)
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_CORE) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_CORE)/$(CORE_BINARYNAME)
	@echo "Done."

compile_jobservice:
	@echo "compiling binary for jobservice (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_JOBSERVICE) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_JOBSERVICE)/$(JOBSERVICEBINARYNAME)
	@echo "Done."

compile_registryctl:
	@echo "compiling binary for harbor registry controller (golang image)..."
	@$(DOCKERCMD) run --rm -v $(BUILDPATH):$(GOBUILDPATH) -w $(GOBUILDPATH_REGISTRYCTL) $(GOBUILDIMAGE) $(GOIMAGEBUILD) -o $(GOBUILDMAKEPATH_REGISTRYCTL)/$(REGISTRYCTLBINARYNAME)
	@echo "Done."

compile:check_environment compile_adminserver compile_core compile_jobservice compile_registryctl
	
prepare:
	@echo "preparing..."
	@$(MAKEPATH)/$(PREPARECMD) $(PREPARECMD_PARA)

build:
	make -f $(MAKEFILEPATH_PHOTON)/Makefile build -e DEVFLAG=$(DEVFLAG) \
	 -e REGISTRYVERSION=$(REGISTRYVERSION) -e NGINXVERSION=$(NGINXVERSION) -e NOTARYVERSION=$(NOTARYVERSION) \
	 -e CLAIRVERSION=$(CLAIRVERSION) -e CLAIRDBVERSION=$(CLAIRDBVERSION) -e VERSIONTAG=$(VERSIONTAG) \
	 -e BUILDBIN=$(BUILDBIN) -e REDISVERSION=$(REDISVERSION) -e MIGRATORVERSION=$(MIGRATORVERSION) \
	 -e CHARTMUSEUMVERSION=$(CHARTMUSEUMVERSION) -e DOCKERIMAGENAME_CHART_SERVER=$(DOCKERIMAGENAME_CHART_SERVER)

modify_composefile: modify_composefile_notary modify_composefile_clair modify_composefile_chartmuseum
	@echo "preparing docker-compose file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSETPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i -e 's/__version__/$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i -e 's/__postgresql_version__/$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i -e 's/__reg_version__/$(REGISTRYVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i -e 's/__nginx_version__/$(NGINXVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)
	@$(SEDCMD) -i -e 's/__redis_version__/$(REDISVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSEFILENAME)

modify_composefile_notary:
	@echo "preparing docker-compose notary file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYTPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)
	@$(SEDCMD) -i -e 's/__notary_version__/$(NOTARYVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSENOTARYFILENAME)

modify_composefile_clair:
	@echo "preparing docker-compose clair file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRTPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
	@$(SEDCMD) -i -e 's/__postgresql_version__/$(CLAIRDBVERSION)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)
	@$(SEDCMD) -i -e 's/__clair_version__/$(CLAIRVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECLAIRFILENAME)

modify_composefile_chartmuseum:
	@echo "preparing docker-compose chartmuseum file..."
	@cp $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECHARTMUSEUMTPLFILENAME) $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECHARTMUSEUMFILENAME)
	@$(SEDCMD) -i -e 's/__chartmuseum_version__/$(CHARTMUSEUMVERSION)-$(VERSIONTAG)/g' $(DOCKERCOMPOSEFILEPATH)/$(DOCKERCOMPOSECHARTMUSEUMFILENAME)

modify_sourcefiles:
	@echo "change mode of source files."
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer.key
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer.crt
	@chmod 600 $(MAKEPATH)/common/templates/notary/notary-signer-ca.crt
	@chmod 600 $(MAKEPATH)/common/templates/core/private_key.pem
	@chmod 600 $(MAKEPATH)/common/templates/registry/root.crt

install: compile ui_version build modify_sourcefiles prepare modify_composefile start

package_online: modify_composefile
	@echo "packing online package ..."
	@cp -r make $(HARBORPKG)
	@if [ -n "$(REGISTRYSERVER)" ] ; then \
		$(SEDCMD) -i -e 's/image\: goharbor/image\: $(REGISTRYSERVER)\/$(REGISTRYPROJECTNAME)/' \
		$(HARBORPKG)/docker-compose.yml ; \
	fi
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp open_source_license $(HARBORPKG)/open_source_license
	
	@$(TARCMD) $(PACKAGE_ONLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."
	
package_offline: compile ui_version build modify_sourcefiles modify_composefile
	@echo "packing offline package ..."
	@cp -r make $(HARBORPKG)
	@cp LICENSE $(HARBORPKG)/LICENSE
	@cp open_source_license $(HARBORPKG)/open_source_license

	@echo "saving harbor docker image"
	@$(DOCKERSAVE) $(DOCKERSAVE_PARA) > $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar
	@gzip $(HARBORPKG)/$(DOCKERIMGFILE).$(VERSIONTAG).tar

	@$(TARCMD) $(PACKAGE_OFFLINE_PARA)
	@rm -rf $(HARBORPKG)
	@echo "Done."

gosec:
	#go get github.com/securego/gosec/cmd/gosec
	#go get github.com/dghubble/sling
	@echo "run secure go scan ..."
	@if [ "$(GOSECRESULTS)" != "" ] ; then \
		$(GOPATH)/bin/gosec -fmt=json -out=$(GOSECRESULTS) -quiet ./... | true ; \
	else \
		$(GOPATH)/bin/gosec -fmt=json -out=harbor_gas_output.json -quiet ./... | true ; \
	fi

go_check: misspell golint govet gofmt commentfmt

gofmt:
	@echo checking gofmt...
	@res=$$(gofmt -d -e -s $$(find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print)); \
	if [ -n "$${res}" ]; then \
		echo checking gofmt fail... ; \
		echo "$${res}"; \
		exit 1; \
	fi

commentfmt:
	@echo checking comment format...
	@res=$$(find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print | xargs egrep '(^|\s)\/\/(\S)'); \
	if [ -n "$${res}" ]; then \
		echo checking comment format fail.. ; \
		echo missing whitespace between // and comment body;\
		echo "$${res}"; \
		exit 1; \
	fi

misspell:
	@echo checking misspell...
	@find . -type d \( -path ./src/vendor -o -path ./tests \) -prune -o -name '*.go' -print | xargs misspell -error

golint:
	@echo checking golint...
	@go list ./... | grep -v -E 'vendor|test' | xargs -L1 fgt golint

govet:
	@echo checking govet...
	@go list ./... | grep -v -E 'vendor|test' | xargs -L1 go vet

pushimage:
	@echo "pushing harbor images ..."
	@$(DOCKERTAG) $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_PORTAL):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_CORE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_CORE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_LOG):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

	@$(DOCKERTAG) $(DOCKERIMAGENAME_DB):$(VERSIONTAG) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	@$(PUSHSCRIPTPATH)/$(PUSHSCRIPTNAME) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG) \
		$(REGISTRYUSER) $(REGISTRYPASSWORD) $(REGISTRYSERVER)
	@$(DOCKERRMIMAGE) $(REGISTRYSERVER)$(DOCKERIMAGENAME_DB):$(VERSIONTAG)

start:
	@echo "loading harbor images..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_LIST) up -d
	@echo "Start complete. You can visit harbor now."

down:
	@echo "Please make sure to set -e NOTARYFLAG=true/CLAIRFLAG=true/CHARTFLAG=true if you are using Notary/CLAIR/Chartmuseum in Harbor, otherwise the Notary/CLAIR/Chartmuseum containers cannot be stop automaticlly."
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Type anything but Y or y to exit. [Y/N]: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)
	@echo "stoping harbor instance..."
	@$(DOCKERCOMPOSECMD) $(DOCKERCOMPOSE_LIST) down -v
	@echo "Done."

swagger_client:
	@echo "Generate swagger client"
	wget -q http://central.maven.org/maven2/io/swagger/swagger-codegen-cli/2.3.1/swagger-codegen-cli-2.3.1.jar -O swagger-codegen-cli.jar
	rm -rf harborclient
	mkdir harborclient
	java -jar swagger-codegen-cli.jar generate -i docs/swagger.yaml -l python -o harborclient
	cd harborclient; python ./setup.py install
	pip install docker -q 
	pip freeze

cleanbinary:
	@echo "cleaning binary..."
	@if [ -f $(ADMINSERVERBINARYPATH)/$(ADMINSERVERBINARYNAME) ] ; then rm $(ADMINSERVERBINARYPATH)/$(ADMINSERVERBINARYNAME) ; fi
	@if [ -f $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ] ; then rm $(CORE_BINARYPATH)/$(CORE_BINARYNAME) ; fi
	@if [ -f $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ] ; then rm $(JOBSERVICEBINARYPATH)/$(JOBSERVICEBINARYNAME) ; fi

cleanimage:
	@echo "cleaning image for photon..."
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_ADMINSERVER):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_CORE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_DB):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_JOBSERVICE):$(VERSIONTAG)
	- $(DOCKERRMIMAGE) -f $(DOCKERIMAGENAME_LOG):$(VERSIONTAG)

cleandockercomposefile:
	@echo "cleaning docker-compose files in $(DOCKERCOMPOSEFILEPATH)"
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml" -exec rm -f {} \;
	@find $(DOCKERCOMPOSEFILEPATH) -maxdepth 1 -name "docker-compose*.yml-e" -exec rm -f {} \;

cleanversiontag:
	@echo "cleaning version TAG"
	@rm -rf $(VERSIONFILEPATH)/$(VERSIONFILENAME)

cleanpackage:
	@echo "cleaning harbor install package"
	@if [ -d $(BUILDPATH)/harbor ] ; then rm -rf $(BUILDPATH)/harbor ; fi
	@if [ -f $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-online-installer-$(VERSIONTAG).tgz ; fi
	@if [ -f $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ] ; \
	then rm $(BUILDPATH)/harbor-offline-installer-$(VERSIONTAG).tgz ; fi

.PHONY: cleanall
cleanall: cleanbinary cleanimage cleandockercomposefile cleanversiontag cleanpackage

clean:
	@echo "  make cleanall:		remove binary, Harbor images, specific version docker-compose"
	@echo "		file, specific version tag, online and offline install package"
	@echo "  make cleanbinary:		remove core and jobservice binary"
	@echo "  make cleanimage:		remove Harbor images"
	@echo "  make cleandockercomposefile:	remove specific version docker-compose"
	@echo "  make cleanversiontag:		cleanpackageremove specific version tag"
	@echo "  make cleanpackage:		remove online and offline install package"

all: install
