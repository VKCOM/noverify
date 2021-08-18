/**
 * WasmConfigurator implements configurations using Wasm.
 */
class WasmConfigurator implements IConfigurator {
  public config: Config
  public wasm: GolangWasm
  /**
   * onChange callback is called every time the config changes.
   * @param config new config
   */
  public onChange: (config: Config) => void

  constructor() {
    this.wasm = GolangWasmSingleton.getInstance()
    this.changeConfig(new Config())
  }

  changeConfig(config: Config) {
    this.config = config

    // We need to change the shared variables so that we can read the new config from Go.
    this.wasm.props.configJson = JSON.stringify(this.config)

    if (this.onChange !== undefined) {
      this.onChange(this.config)
    }
  }

  getConfig(): Config {
    return this.config
  }
}
