class PlaygroundStorage {
  public saveCode(editor: CodeMirror.Editor): void {
    if (!editor) {
      return;
    }
    localStorage.setItem('noverify-playground-code', editor.getValue());
  }

  public getCode(): string {
    return localStorage.getItem('noverify-playground-code')
  }
}
