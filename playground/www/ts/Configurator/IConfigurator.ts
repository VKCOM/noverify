/**
 * IConfigurator interface is responsible for configuring the Playground
 * without depending on the actual implementation.
 */
interface IConfigurator {
  config: Config
  /**
   * onChange callback is called every time the config changes.
   * @param config
   */
  onChange: (config: Config) => void

  getConfig(): Config
  changeConfig(config: Config)
}
