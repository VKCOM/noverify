"use strict";

// This is the modified extension.js from https://github.com/felixfbecker/php-language-server

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const path = require("path");
const child_process_1 = require("mz/child_process");
const vscode = require("vscode");
const vscode_languageclient_1 = require("vscode-languageclient");
const semver = require("semver");
const net = require("net");
const url = require("url");
function activate(context) {
    return __awaiter(this, void 0, void 0, function* () {
        const conf = vscode.workspace.getConfiguration('php');
        const executablePath = conf.get('executablePath') || 'php';
        const memoryLimit = conf.get('memoryLimit') || '4095M';
        if (memoryLimit !== '-1' && !/^\d+[KMG]?$/.exec(memoryLimit)) {
            const selected = yield vscode.window.showErrorMessage('The memory limit you\'d provided is not numeric, nor "-1" nor valid php shorthand notation!', 'Open settings');
            if (selected === 'Open settings') {
                yield vscode.commands.executeCommand('workbench.action.openGlobalSettings');
            }
            return;
        }
        let stdout;
        try {
            [stdout] = yield child_process_1.execFile(executablePath, ['--version']);
        }
        catch (err) {
            if (err.code === 'ENOENT') {
                const selected = yield vscode.window.showErrorMessage('PHP executable not found. Install PHP 7 and add it to your PATH or set the php.executablePath setting', 'Open settings');
                if (selected === 'Open settings') {
                    yield vscode.commands.executeCommand('workbench.action.openGlobalSettings');
                }
            }
            else {
                vscode.window.showErrorMessage('Error spawning PHP: ' + err.message);
                console.error(err);
            }
            return;
        }
        const match = stdout.match(/^PHP ([^\s]+)/m);
        if (!match) {
            vscode.window.showErrorMessage('Error parsing PHP version. Please check the output of php --version');
            return;
        }
        let version = match[1].split('-')[0];
        if (!/^\d+.\d+.\d+$/.test(version)) {
            version = version.replace(/(\d+.\d+.\d+)/, '$1-');
        }
        if (semver.lt(version, '7.0.0')) {
            vscode.window.showErrorMessage('The language server needs at least PHP 7 installed. Version found: ' + version);
            return;
        }

        const clientOptions = {
            documentSelector: [
                { scheme: 'file', language: 'php' },
                { scheme: 'untitled', language: 'php' }
            ],
            uriConverters: {
                code2Protocol: uri => url.format(url.parse(uri.toString(true))),
                protocol2Code: str => vscode.Uri.parse(str)
            },
            synchronize: {
                configurationSection: 'php',
                fileEvents: vscode.workspace.createFileSystemWatcher('**/*.php')
            }
        };
        const disposable = new vscode_languageclient_1.LanguageClient(
            'PHP Language Server',
            {
                command: '/path/to/noverify',
                args: ['-cores=4', '-lang-server', '-cache-dir=/path/to/cache', '-stubs-dir=/path/to/phpstorm-stubs']
            },
            clientOptions
        ).start();
        context.subscriptions.push(disposable);
    });
}
exports.activate = activate;
