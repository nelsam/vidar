
# Create the build directory.
build:
	@mkdir build

# Build the goimports plugin.
build/goimports.so: | build
	go build -buildmode plugin -o ./build/goimports.so github.com/nelsam/vidar/plugin/goimports/main

# Build the comments plugin.
build/comments.so: | build
	go build -buildmode plugin -o ./build/comments.so github.com/nelsam/vidar/plugin/comments/main

# Build the godef plugin.
build/godef.so: | build
	go build -buildmode plugin -o ./build/godef.so github.com/nelsam/vidar/plugin/godef/main

# Build the license plugin.
build/license.so: | build
	go build -buildmode plugin -o ./build/license.so github.com/nelsam/vidar/plugin/license/main

# Build all plugins included with vidar.
plugins: build/goimports.so build/comments.so build/godef.so build/license.so
.PHONY: plugins

# Install all plugins included with vidar to
# $HOME/.config/vidar/plugins.  If plugins have
# not yet been built, they will be built before
# installing.
plugins-install: plugins
	@mkdir -p $$HOME/.config/vidar/plugins
	@for plugin in $$(ls -1 build); do \
		echo "Installing $$plugin to $$HOME/.config/vidar/plugins"; \
		install build/$$plugin $$HOME/.config/vidar/plugins/; \
	done
.PHONY: plugins-install
	
clean:
	@rm -rf build
.PHONY: clean
