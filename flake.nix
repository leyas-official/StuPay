{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs?tag=25.05";
    systems.url = "github:nix-systems/default";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      treefmt-nix,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; };

      treefmtEval = treefmt-nix.lib.evalModule pkgs ./treefmt.nix;

      uniContentApi = pkgs.buildGoModule {
        pname = "uni-content-api";
        version = builtins.substring 0 8 (self.lastModifiedDate or "19700101");
        src = self.outPath;
        vendorHash = "sha256-fbB01/Jj20VsKkLVBtEEahFYi7ZMKF4pJxlwtUoLYMo=";
        meta = { };
        doCheck = false;
      };

      dockerImage = pkgs.dockerTools.buildImage {
        name = "uni-content-api";
        tag = "latest";
        copyToRoot = [ uniContentApi ];
        config = { 
          Cmd = [ "${uniContentApi}/bin/UniContentAPI" ]; 
        };
      };

      addAdmin = pkgs.writeShellScriptBin "addAdmin" ''
        USER_NAME=$(${pkgs.gum}/bin/gum input --placeholder "user name")
        if [[ "$USER_NAME" == "" ]]; then
          echo "User name cannot be empty"
          exit 1
        fi
        EMAIL=$(${pkgs.gum}/bin/gum input --placeholder "email")
        if [[ "$EMAIL" == "" ]]; then
          echo "Email cannot be empty"
          exit 1
        fi
        PASSWORD=$(${pkgs.gum}/bin/gum input --placeholder "password")
        if [[ "$PASSWORD" == "" ]]; then
          echo "Password cannot be empty"
          exit 1
        fi
        JSON=$(${pkgs.jq}/bin/jq -n --arg user_name "$USER_NAME" --arg email "$EMAIL" --arg password "$PASSWORD" '{"Username": $user_name, "Email": $email, "Password": $password, "PasswordConfirm": $password}')

        ${pkgs.curl}/bin/curl -X POST -H "Content-Type: application/json" -d "'$JSON'" http://localhost:8088/auth

        stmt="INSERT INTO user_roles (user_id, role_id) VALUES ((SELECT id FROM api_user WHERE user_name = '$USER_NAME'), (SELECT id FROM role WHERE name = 'admin'))"

        ${pkgs.postgresql}/bin/psql -d "$GOOSE_DBSTRING" -c "$stmt"
      '';
    in
    {
      # Build executables. See https://nixos.org/manual/nixpkgs/stable/#sec-language-go
      packages = {
        inherit uniContentApi dockerImage;
        default = uniContentApi;
      };

      devShells = {
        default = pkgs.mkShell {
        DOCKER_HOST = "unix:///run/user/1000/podman/podman.sock";
        TESTCONTAINERS_RYUK_DISABLED="true";
        TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE = "unix:///run/user/1000/podman/podman.sock";
          packages = [
              addAdmin
            ] ++ (with pkgs; [
            just

            go
            gopls
            goose
            air
            postgresql
            jq
          ]);
        };
      };

      # Provides a script that copies required files to ~/
      podmanSetupScript = 
        let
          registriesConf = pkgs.writeText "registries.conf" ''
            [registries.search]
            registries = ['docker.io']
            [registries.block]
            registries = []
          '';
        in
        pkgs.writeScript "podman-setup" ''
          #!${pkgs.runtimeShell}
          # Dont overwrite customised configuration
          if ! test -f ~/.config/containers/policy.json; then
            install -Dm555 ${pkgs.skopeo.src}/default-policy.json ~/.config/containers/policy.json
          fi
          if ! test -f ~/.config/containers/registries.conf; then
            install -Dm555 ${registriesConf} ~/.config/containers/registries.conf
          fi
        '';

      dockerCompat = pkgs.runCommandNoCC "docker-podman-compat" { } ''
          mkdir -p $out/bin
          ln -s ${pkgs.podman}/bin/podman $out/bin/docker
        '';

      formatter = treefmtEval.${pkgs.system}.config.build.wrapper;

      checks = {
        treefmt = treefmtEval.${pkgs.system}.config.build.check self;
      };
    });
}
