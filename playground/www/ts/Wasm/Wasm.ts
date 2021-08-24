class GolangWasmSingleton {
  private static instance: GolangWasm;

  protected constructor() {}

  public static getInstance(): GolangWasm {
    if (!GolangWasmSingleton.instance) {
      GolangWasmSingleton.instance = new GolangWasm();
    }
    return GolangWasmSingleton.instance;
  }
}

class GolangWasmProperties {
  public analyzeCallback: () => void
  public reportsJson: string = '[]'
  public configJson: string = '{}'
}

/**
 * The GolangWasm class is responsible for starting a stream for wasm,
 * as well as storing fields that will be passed from JS to Golang and
 * vice versa.
 */
class GolangWasm {
  private go: any
  public props: GolangWasmProperties

  constructor() {
    // @ts-ignore
    this.go = new Go()
    this.props = new GolangWasmProperties()
  }

  public run(callback: () => void) {
    WebAssembly.instantiateStreaming(fetch('main.wasm'), this.go.importObject).then((result) => {
      this.go.run(result.instance)
      callback()
    });
  }
}
