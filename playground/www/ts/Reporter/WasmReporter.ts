/**
 * WasmReporter is class implements receiving reports using Wasm.
 */
class WasmReporter implements IReporter {
  private wasm: GolangWasm

  constructor() {
    this.wasm = GolangWasmSingleton.getInstance()
  }

  public getReports(code: string): Report[] {
    if (this.wasm.props.analyzeCallback !== undefined) {
      this.wasm.props.analyzeCallback()
    }

    let reports: Report[]

    try {
      reports = JSON.parse(this.wasm.props.reportsJson)
    } catch (e) {
      console.error("Error parse reports json: ", this.wasm.props.reportsJson)
      return []
    }

    if (reports === null) {
      return []
    }

    return reports
  }
}
