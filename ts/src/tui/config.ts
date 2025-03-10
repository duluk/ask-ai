// This is a wrapper file for ESM imports
import { loadConfig as originalLoadConfig, getModelName, getDefaultModel } from '../config/index.js';

export {
  getModelName,
  getDefaultModel
};

export const loadConfig = originalLoadConfig;