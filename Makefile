
# Create the build directory.
build:
	mkdir build

# Create the plugin install directory.
$(HOME)/.local/share/vidar/plugins:
	mkdir -p $(HOME)/.local/share/vidar/plugins

# Build the gosyntax plugin.
build/gosyntax.so: | build
	go build -buildmode plugin -o ./build/gosyntax.so github.com/nelsam/vidar/plugin/go/gosyntax/main

# Build the goimports plugin.
build/goimports.so: | build
	go build -buildmode plugin -o ./build/goimports.so github.com/nelsam/vidar/plugin/go/goimports/main

# Build the comments plugin.
build/comments.so: | build
	go build -buildmode plugin -o ./build/comments.so github.com/nelsam/vidar/plugin/go/comments/main

# Build the godef plugin.
build/godef.so: | build
	go build -buildmode plugin -o ./build/godef.so github.com/nelsam/vidar/plugin/go/godef/main

# Build the goguru plugin.
build/goguru.so: | build
	go build -buildmode plugin -o ./build/goguru.so github.com/nelsam/vidar/plugin/go/goguru/main

# Build the gocode plugin.
build/gocode.so: | build
	go build -buildmode plugin -o ./build/gocode.so github.com/nelsam/vidar/plugin/go/gocode/main

# Build the license plugin.
build/license.so: | build
	go build -buildmode plugin -o ./build/license.so github.com/nelsam/vidar/plugin/go/license/main

# Build all plugins included with vidar.
#
# Note: this currently chooses godef.so over goguru.so, since they have overlapping functionality and guru
# support is new (and still experimental).
plugins: build/gosyntax.so build/goimports.so build/comments.so build/godef.so build/license.so build/gocode.so
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
