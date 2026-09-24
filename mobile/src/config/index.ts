import { Platform } from 'react-native';

const raw = (process.env.EXPO_PUBLIC_API_URL ?? '').trim();
const fallback = Platform.select({
  android: 'http://10.0.2.2:3000',
  default: 'http://localhost:3000',
});

export const API_BASE_URL = (raw || fallback).replace(/\/+$/, '');
