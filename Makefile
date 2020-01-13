define depslist =
	$(shell go list -f '{{join .Deps " "}}' $(1))
endef

define depsfiles =
	$(foreach p,$(call depslist,$(1)),$(wildcard $(GOPATH)/src/$(p)/*.go))
endef

# Create the build directory.
build:
	mkdir build

# Create the plugin install directory.
$(HOME)/.local/share/vidar/plugins:
	mkdir -p $(HOME)/.local/share/vidar/plugins

# Build the gosyntax plugin.
build/gosyntax.so: $(call depsfiles,github.com/nelsam/vidar/plugin/gosyntax/main) | build
	go build -buildmode plugin -o ./build/gosyntax.so github.com/nelsam/vidar/plugin/gosyntax/main

# Build the goimports plugin.
build/goimports.so: $(call depsfiles,github.com/nelsam/vidar/plugin/goimports/main) | build
	go build -buildmode plugin -o ./build/goimports.so github.com/nelsam/vidar/plugin/goimports/main

# Build the comments plugin.
build/comments.so: $(call depsfiles,github.com/nelsam/vidar/plugin/comments/main) | build
	go build -buildmode plugin -o ./build/comments.so github.com/nelsam/vidar/plugin/comments/main

# Build the godef plugin.
build/godef.so: $(call depsfiles,github.com/nelsam/vidar/plugin/godef/main) | build
	go build -buildmode plugin -o ./build/godef.so github.com/nelsam/vidar/plugin/godef/main

# Build the gocode plugin.
build/gocode.so: $(call depsfiles,github.com/nelsam/vidar/plugin/gocode/main) | build
	go build -buildmode plugin -o ./build/gocode.so github.com/nelsam/vidar/plugin/gocode/main

# Build the license plugin.
build/license.so: $(call depsfiles,github.com/nelsam/vidar/plugin/license/main) | build
	go build -buildmode plugin -o ./build/license.so github.com/nelsam/vidar/plugin/license/main

# Build the vsyntax plugin.
build/vsyntax.so: $(call depsfiles,github.com/nelsam/vidar/plugin/vsyntax/main) | build
	go build -buildmode plugin -o ./build/vsyntax.so github.com/nelsam/vidar/plugin/vsyntax/main

# Build the vfmt plugin.
build/vfmt.so: $(call depsfiles,github.com/nelsam/vidar/plugin/vfmt/main) | build
	go build -buildmode plugin -o ./build/vfmt.so github.com/nelsam/vidar/plugin/vfmt/main

# Build all plugins included with vidar.
plugins: build/gosyntax.so build/goimports.so build/comments.so build/godef.so build/license.so build/gocode.so build/vsyntax.so build/vfmt.so
.PHONY: plugins

# Install all plugins included with vidar to
# $HOME/.local/share/vidar/plugins.  If plugins
# have not yet been built, they will be built
# before installing.
plugins-install: plugins | $(HOME)/.local/share/vidar/plugins
	@for plugin in $$(ls -1 build); do \
		echo "Installing $$plugin to $$HOME/.local/share/vidar/plugins"; \
		install build/$$plugin $$HOME/.local/share/vidar/plugins/; \
	done
.PHONY: plugins-install

clean:
	@rm -rf build
.PHONY: clean
