import { register } from 'module';
import * as path from 'path';
import * as fs from 'fs';

function _hasFolderPackageJsonTypeModule(folder) {
  if (folder.endsWith('/node_modules')) {
    return false;
  }
  let pj = path.join(folder, '/package.json');
  if (fs.existsSync(pj)) {
    try {
      let pkg = JSON.parse(fs.readFileSync(pj).toString());
      if (pkg) {
        if (pkg.type === 'module') {
          return true;
        } else {
          return false;
        }
      }
    } catch (e) {
      console.warn(`${pj} cannot be read, it will be ignored for ES module detection purposes.`, e);
      return false;
    }
  }
  if (folder === '/') {
    return false;
  }
  return _hasFolderPackageJsonTypeModule(path.resolve(folder, '..'));
}

function _hasPackageJsonTypeModule(file) {
  let jsPath = file + '.js';
  if (fs.existsSync(jsPath)) {
    return _hasFolderPackageJsonTypeModule(path.resolve(path.dirname(jsPath)));
  }
  return false;
}

function _resolveHandlerFileName() {
  const taskRoot = process.env.LAMBDA_TASK_ROOT;
  const handlerDef = process.env._HANDLER;
  if (!taskRoot || !handlerDef) {
    return null;
  }
  const handler = path.basename(handlerDef);
  const moduleRoot = handlerDef.substr(0, handlerDef.length - handler.length);
  const [module, _] = handler.split('.', 2);
  return path.resolve(taskRoot, moduleRoot, module);
}

function _isHandlerAnESModule() {
  try {
    const handlerFileName = _resolveHandlerFileName();
    if (!handlerFileName) {
      return false;
    }
    if (fs.existsSync(handlerFileName + '.mjs')) {
      return true;
    } else if (fs.existsSync(handlerFileName + '.cjs')) {
      return false;
    } else {
      return _hasPackageJsonTypeModule(handlerFileName);
    }
  } catch (e) {
    console.error('Unknown error occurred while checking whether handler is an ES module', e);
    return false;
  }
}

let registered = false;

export function registerLoader() {
  if (!registered) {
    register('import-in-the-middle/hook.mjs', import.meta.url);
    registered = true;
  }
}

if (_isHandlerAnESModule()) {
  /*
  We could activate ESM loader hook of the "import-in-the-middle" library,
  - by "--loader=import-in-the-middle/hook.mjs" Node CLI option, but "--loader" option has been deprecated
  - or by "--import=import-in-the-middle/hook.mjs" Node CLI option, but in this case,
    there will always be "import-in-the-middle" hook initialization overhead even for non-ESM (CommonJS) modules

  Hence, instead, we initialize "import-in-the-middle" hook only for ES (EcmaScript) based user handlers
  to prevent redundant "import-in-the-middle" hook initialization overhead during coldstart
  of the CommonJS based user handlers.
   */
  registerLoader();
}
