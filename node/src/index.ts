import {
  dag,
  Container,
  Directory,
  object,
  func,
  field,
  CacheVolume,
  Secret,
} from "@dagger.io/dagger"

import { Commands } from "./commands"

@object()
class Node {
  @field()
  version = "18-alpine"

  @field()
  container: Container

  constructor(version?: string, ctr?: Container) {
    this.version = version ?? this.version
    this.container = ctr ?? dag.container().from(`node:${this.version}`)
  }

  /**
   * Add source to the module container.
   *
   * @param source The source directory to mount in the container.
   * @param cache The cache to use for the node_modules cache (default to "node-modules").
   */
  @func()
  withSource(source: Directory, cache?: CacheVolume): Node {
    const workdir = "/src"

    this.container = this.container
      .withWorkdir(workdir)
      .withDirectory(workdir, source)
      .withMountedCache(
        `${workdir}/node_modules`,
        cache ?? dag.cacheVolume("node-modules"),
      )

    return this
  }
  /**
   * Add npmrc to the module container.
   *
   * @param token The npmrc file to mount in the container.
   *
   */
  @func()
  withNpmrc(token: Secret, cache?: CacheVolume): Node {
      dag.container()
      .from("alpine")
      .withSecretVariable("TEST_ENV", token)
      .withEntrypoint([])
      .withExec([
        "sh",
        "-c",
        `echo $TEST_ENV > ~/.npmtest`,
      ])
      .withMountedCache(
        "/root/.npmtest",
        // cache ?? dag.cacheVolume
        cache ?? dag.cacheVolume("~/.npmtest")
      )

    return this
  }


  /**
   * Add npm as package manager in the container.
   *
   * This also update the container entrypoint to "npm".
   * @param cache The cache to use for the downloaded packages (default to "node-module-npm").
   */
  @func()
  withNpm(cache?: CacheVolume): Node {
    this.container = this.container
      .withEntrypoint(["npm"])
      .withMountedCache(
        "/root/.npm",
        cache ?? dag.cacheVolume(`node-module-npm`),
      )

    return this
  }

  /**
   * Downloads dependencies in the container.
   *
   * @param pkgs Additional packages to install in the container.
   */
  @func()
  install(pkgs: string[] = []): Node {
    this.container = this.container.withExec(["install", ...pkgs])

    return this
  }

  /**
   * Execute commands in the container.
   */
  @func()
  commands(): Commands {
    return new Commands(this.container)
  }
}
