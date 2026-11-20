import {
  context,
  ContextManager,
  diag,
  DiagConsoleLogger,
  logs,
  metrics,
  propagation,
  TextMapPropagator,
  trace,
} from '@opentelemetry/api';
import {
  diagLogLevelFromString,
  getNumberFromEnv,
  getStringFromEnv,
  getStringListFromEnv,
} from '@opentelemetry/core';
import { getPropagator } from '@opentelemetry/auto-configuration-propagators';
import { NodeSDK, type NodeSDKConfiguration } from '@opentelemetry/sdk-node';
import type { TracerProviderOptions } from '@opentelemetry/sdk-trace';
import type { MetricReader } from '@opentelemetry/sdk-metrics';
import { envDetector, processDetector } from '@opentelemetry/resources';
import { awsLambdaDetector } from '@opentelemetry/resource-detector-aws';
import {
  Instrumentation,
  registerInstrumentations,
} from '@opentelemetry/instrumentation';
import {
  AwsInstrumentation,
  AwsSdkInstrumentationConfig,
} from '@opentelemetry/instrumentation-aws-sdk';
import {
  AwsLambdaInstrumentation,
  AwsLambdaInstrumentationConfig,
} from '@opentelemetry/instrumentation-aws-lambda';

const defaultInstrumentationList = ['dns', 'http', 'net'];

export interface SDKRegistrationConfig {
  propagator?: TextMapPropagator | null;
  contextManager?: ContextManager | null;
}

// Instrumentations are left out: NodeSDK registers them before it sets up the
// providers, so the layer registers its own after start() instead.
export type SdkConfig = Partial<Omit<NodeSDKConfiguration, 'instrumentations'>>;

declare global {
  // In case of downstream configuring span processors etc
  function configureLambdaInstrumentation(
    config: AwsLambdaInstrumentationConfig,
  ): AwsLambdaInstrumentationConfig;
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
  function configureInstrumentations(): Instrumentation[];

  // Gets the Lambda defaults before the deprecated hooks below update them.
  function configureSdk(defaultConfig: SdkConfig): SdkConfig;

  /** @deprecated Use configureSdk. */
  function configureTracer(
    defaultConfig: TracerProviderOptions,
  ): TracerProviderOptions;
  /** @deprecated Use configureSdk. */
  function configureSdkRegistration(
    defaultSdkRegistration: SDKRegistrationConfig,
  ): SDKRegistrationConfig;

  // No explicit metric type here, but "unknown" type.
  // Because metric packages are important dynamically.
  /** @deprecated Use configureSdk. */
  function configureMeter(defaultConfig: unknown): unknown;

  // No explicit logger type here, but "unknown" type.
  // Because logger packages are important dynamically.
  /** @deprecated Use configureSdk. */
  function configureLogger(defaultConfig: unknown): unknown;
}

function getActiveInstrumentations(): Set<string> {
  let enabledInstrumentations: string[] = defaultInstrumentationList;
  if (process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS) {
    enabledInstrumentations =
      process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS.split(',').map(i =>
        i.trim(),
      );
  }
  const instrumentationSet = new Set<string>(enabledInstrumentations);
  if (process.env.OTEL_NODE_DISABLED_INSTRUMENTATIONS) {
    const disableInstrumentations =
      process.env.OTEL_NODE_DISABLED_INSTRUMENTATIONS.split(',').map(i =>
        i.trim(),
      );
    disableInstrumentations.forEach(di => instrumentationSet.delete(di));
  }
  return instrumentationSet;
}

async function defaultConfigureInstrumentations() {
  const instrumentations = [];
  const activeInstrumentations = getActiveInstrumentations();
  if (activeInstrumentations.has('amqplib')) {
    const { AmqplibInstrumentation } =
      await import('@opentelemetry/instrumentation-amqplib');
    instrumentations.push(new AmqplibInstrumentation());
  }
  if (activeInstrumentations.has('bunyan')) {
    const { BunyanInstrumentation } =
      await import('@opentelemetry/instrumentation-bunyan');
    instrumentations.push(new BunyanInstrumentation());
  }
  if (activeInstrumentations.has('cassandra-driver')) {
    const { CassandraDriverInstrumentation } =
      await import('@opentelemetry/instrumentation-cassandra-driver');
    instrumentations.push(new CassandraDriverInstrumentation());
  }
  if (activeInstrumentations.has('connect')) {
    const { ConnectInstrumentation } =
      await import('@opentelemetry/instrumentation-connect');
    instrumentations.push(new ConnectInstrumentation());
  }
  if (activeInstrumentations.has('dataloader')) {
    const { DataloaderInstrumentation } =
      await import('@opentelemetry/instrumentation-dataloader');
    instrumentations.push(new DataloaderInstrumentation());
  }
  if (activeInstrumentations.has('dns')) {
    const { DnsInstrumentation } =
      await import('@opentelemetry/instrumentation-dns');
    instrumentations.push(new DnsInstrumentation());
  }
  if (activeInstrumentations.has('express')) {
    const { ExpressInstrumentation } =
      await import('@opentelemetry/instrumentation-express');
    instrumentations.push(new ExpressInstrumentation());
  }
  if (activeInstrumentations.has('fs')) {
    const { FsInstrumentation } =
      await import('@opentelemetry/instrumentation-fs');
    instrumentations.push(new FsInstrumentation());
  }
  if (activeInstrumentations.has('graphql')) {
    const { GraphQLInstrumentation } =
      await import('@opentelemetry/instrumentation-graphql');
    instrumentations.push(new GraphQLInstrumentation());
  }
  if (activeInstrumentations.has('grpc')) {
    const { GrpcInstrumentation } =
      await import('@opentelemetry/instrumentation-grpc');
    instrumentations.push(new GrpcInstrumentation());
  }
  if (activeInstrumentations.has('hapi')) {
    const { HapiInstrumentation } =
      await import('@opentelemetry/instrumentation-hapi');
    instrumentations.push(new HapiInstrumentation());
  }
  if (activeInstrumentations.has('http')) {
    const { HttpInstrumentation } =
      await import('@opentelemetry/instrumentation-http');
    instrumentations.push(new HttpInstrumentation());
  }
  if (activeInstrumentations.has('ioredis')) {
    const { IORedisInstrumentation } =
      await import('@opentelemetry/instrumentation-ioredis');
    instrumentations.push(new IORedisInstrumentation());
  }
  if (activeInstrumentations.has('kafkajs')) {
    const { KafkaJsInstrumentation } =
      await import('@opentelemetry/instrumentation-kafkajs');
    instrumentations.push(new KafkaJsInstrumentation());
  }
  if (activeInstrumentations.has('knex')) {
    const { KnexInstrumentation } =
      await import('@opentelemetry/instrumentation-knex');
    instrumentations.push(new KnexInstrumentation());
  }
  if (activeInstrumentations.has('koa')) {
    const { KoaInstrumentation } =
      await import('@opentelemetry/instrumentation-koa');
    instrumentations.push(new KoaInstrumentation());
  }
  if (activeInstrumentations.has('memcached')) {
    const { MemcachedInstrumentation } =
      await import('@opentelemetry/instrumentation-memcached');
    instrumentations.push(new MemcachedInstrumentation());
  }
  if (activeInstrumentations.has('mongodb')) {
    const { MongoDBInstrumentation } =
      await import('@opentelemetry/instrumentation-mongodb');
    instrumentations.push(new MongoDBInstrumentation());
  }
  if (activeInstrumentations.has('mongoose')) {
    const { MongooseInstrumentation } =
      await import('@opentelemetry/instrumentation-mongoose');
    instrumentations.push(new MongooseInstrumentation());
  }
  if (activeInstrumentations.has('mysql')) {
    const { MySQLInstrumentation } =
      await import('@opentelemetry/instrumentation-mysql');
    instrumentations.push(new MySQLInstrumentation());
  }
  if (activeInstrumentations.has('mysql2')) {
    const { MySQL2Instrumentation } =
      await import('@opentelemetry/instrumentation-mysql2');
    instrumentations.push(new MySQL2Instrumentation());
  }
  if (activeInstrumentations.has('nestjs-core')) {
    const { NestInstrumentation } =
      await import('@opentelemetry/instrumentation-nestjs-core');
    instrumentations.push(new NestInstrumentation());
  }
  if (activeInstrumentations.has('net')) {
    const { NetInstrumentation } =
      await import('@opentelemetry/instrumentation-net');
    instrumentations.push(new NetInstrumentation());
  }
  if (activeInstrumentations.has('pg')) {
    const { PgInstrumentation } =
      await import('@opentelemetry/instrumentation-pg');
    instrumentations.push(new PgInstrumentation());
  }
  if (activeInstrumentations.has('pino')) {
    const { PinoInstrumentation } =
      await import('@opentelemetry/instrumentation-pino');
    instrumentations.push(new PinoInstrumentation());
  }
  if (activeInstrumentations.has('redis')) {
    const { RedisInstrumentation } =
      await import('@opentelemetry/instrumentation-redis');
    instrumentations.push(new RedisInstrumentation());
  }
  if (activeInstrumentations.has('restify')) {
    const { RestifyInstrumentation } =
      await import('@opentelemetry/instrumentation-restify');
    instrumentations.push(new RestifyInstrumentation());
  }
  if (activeInstrumentations.has('socket.io')) {
    const { SocketIoInstrumentation } =
      await import('@opentelemetry/instrumentation-socket.io');
    instrumentations.push(new SocketIoInstrumentation());
  }
  if (activeInstrumentations.has('undici')) {
    const { UndiciInstrumentation } =
      await import('@opentelemetry/instrumentation-undici');
    instrumentations.push(new UndiciInstrumentation());
  }
  if (activeInstrumentations.has('winston')) {
    const { WinstonInstrumentation } =
      await import('@opentelemetry/instrumentation-winston');
    instrumentations.push(new WinstonInstrumentation());
  }
  return instrumentations;
}

async function createInstrumentations() {
  return [
    new AwsInstrumentation(
      typeof configureAwsInstrumentation === 'function'
        ? configureAwsInstrumentation({ suppressInternalInstrumentation: true })
        : { suppressInternalInstrumentation: true },
    ),
    new AwsLambdaInstrumentation(
      typeof configureLambdaInstrumentation === 'function'
        ? configureLambdaInstrumentation({})
        : {},
    ),
    ...(typeof configureInstrumentations === 'function'
      ? configureInstrumentations()
      : await defaultConfigureInstrumentations()),
  ];
}

async function initializeSdk() {
  // Read while the layer logger is active so invalid propagator warnings show.
  let sdkConfig: SdkConfig = {
    autoDetectResources: true,
    resourceDetectors: [awsLambdaDetector, envDetector, processDetector],
    textMapPropagator: getPropagator(),
  };
  if (typeof configureSdk === 'function') {
    sdkConfig = configureSdk(sdkConfig);
  }

  // Deprecated hooks update the SDK options only when defined.
  if (typeof configureTracer === 'function') {
    const tracerConfig = configureTracer({
      resource: sdkConfig.resource,
      sampler: sdkConfig.sampler,
      spanLimits: sdkConfig.spanLimits,
      idGenerator: sdkConfig.idGenerator,
      spanProcessors: sdkConfig.spanProcessors ?? [],
    });
    if (Object.hasOwn(tracerConfig, 'resource')) {
      sdkConfig.resource = tracerConfig.resource;
    }
    if (Object.hasOwn(tracerConfig, 'sampler')) {
      sdkConfig.sampler = tracerConfig.sampler;
    }
    if (Object.hasOwn(tracerConfig, 'spanLimits')) {
      sdkConfig.spanLimits = tracerConfig.spanLimits;
    }
    if (Object.hasOwn(tracerConfig, 'idGenerator')) {
      sdkConfig.idGenerator = tracerConfig.idGenerator;
    }
    if (Object.hasOwn(tracerConfig, 'spanProcessors')) {
      sdkConfig.spanProcessors =
        sdkConfig.spanProcessors === undefined
          ? nonEmpty(tracerConfig.spanProcessors)
          : tracerConfig.spanProcessors;
    }
  }

  // Only readers and views are supported from the deprecated meter hook.
  if (typeof configureMeter === 'function') {
    // The hook gets the env readers, as before NodeSDK, so a hook that adds a
    // reader keeps the default OTLP export.
    const defaults = sdkConfig.metricReaders
      ? undefined
      : await metricReadersFromEnv();
    const readers = sdkConfig.metricReaders ?? [...(defaults ?? [])];
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const meterConfig: any =
      configureMeter({
        readers,
        views: sdkConfig.views,
      }) ?? {};
    if (Object.hasOwn(meterConfig, 'readers')) {
      // Left unchanged, NodeSDK builds the env readers itself.
      sdkConfig.metricReaders =
        defaults &&
        meterConfig.readers === readers &&
        sameItems(readers, defaults)
          ? undefined
          : meterConfig.readers;
    }
    if (Object.hasOwn(meterConfig, 'views')) {
      sdkConfig.views = meterConfig.views;
    }
  }

  // Only processors are supported from the deprecated logger hook.
  if (typeof configureLogger === 'function') {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const loggerConfig: any =
      configureLogger({
        processors: sdkConfig.logRecordProcessors ?? [],
      }) ?? {};
    if (Object.hasOwn(loggerConfig, 'processors')) {
      sdkConfig.logRecordProcessors =
        sdkConfig.logRecordProcessors === undefined
          ? nonEmpty(loggerConfig.processors)
          : loggerConfig.processors;
    }
  }

  if (typeof configureSdkRegistration === 'function') {
    const sdkRegistrationConfig = configureSdkRegistration({
      propagator: sdkConfig.textMapPropagator,
      contextManager: sdkConfig.contextManager,
    });
    if (Object.hasOwn(sdkRegistrationConfig, 'propagator')) {
      sdkConfig.textMapPropagator =
        sdkRegistrationConfig.propagator ?? sdkConfig.textMapPropagator;
    }
    if (Object.hasOwn(sdkRegistrationConfig, 'contextManager')) {
      // NodeSDK treats null as "leave the context manager unset", as register() did.
      sdkConfig.contextManager =
        sdkRegistrationConfig.contextManager as ContextManager;
    }
  }

  // NodeSDK installs its own diag logger when OTEL_LOG_LEVEL is set.
  // Drop ours first so it does not warn about the override.
  if (getStringFromEnv('OTEL_LOG_LEVEL') != null) {
    diag.disable();
  }

  const sdk = new NodeSDK(sdkConfig);
  sdk.start();

  // Create instrumentations if they have not been created before
  // to prevent additional coldstart overhead
  // caused by creations and initializations of instrumentations.
  if (!instrumentations || !instrumentations.length) {
    instrumentations = await createInstrumentations();
  }

  // Register the instrumentations again now that NodeSDK has set the global
  // providers. They are not passed to NodeSDK: start() registers them before
  // it creates the providers, so the Lambda instrumentation would bind the
  // no-op tracer provider and never flush at the end of an invocation.
  disableInstrumentations = registerInstrumentations({ instrumentations });
}

function nonEmpty<T>(list: T[] | undefined): T[] | undefined {
  return list && list.length ? list : undefined;
}

function sameItems<T>(list: T[], other: T[]): boolean {
  return list.length === other.length && list.every((v, i) => v === other[i]);
}

// The readers NodeSDK builds from OTEL_METRICS_EXPORTER. Prometheus is left
// out: its exporter starts a server as soon as it is created.
async function metricReadersFromEnv(): Promise<MetricReader[]> {
  const exporters = new Set(getStringListFromEnv('OTEL_METRICS_EXPORTER'));
  if (exporters.has('none')) {
    return [];
  }
  if (exporters.size === 0) {
    exporters.add('otlp');
  }
  const { ConsoleMetricExporter, PeriodicExportingMetricReader } =
    await import('@opentelemetry/sdk-metrics');
  const readers: MetricReader[] = [];
  if (exporters.has('otlp')) {
    const interval = positiveFromEnv('OTEL_METRIC_EXPORT_INTERVAL');
    const timeout = positiveFromEnv('OTEL_METRIC_EXPORT_TIMEOUT');
    readers.push(
      new PeriodicExportingMetricReader({
        exporter: await otlpMetricExporterFromEnv(),
        ...(interval && { exportIntervalMillis: interval }),
        ...(timeout && {
          exportTimeoutMillis: interval ? Math.min(timeout, interval) : timeout,
        }),
      }),
    );
  }
  if (exporters.has('console')) {
    readers.push(
      new PeriodicExportingMetricReader({
        exporter: new ConsoleMetricExporter(),
      }),
    );
  }
  return readers;
}

async function otlpMetricExporterFromEnv() {
  const protocol = (
    getStringFromEnv('OTEL_EXPORTER_OTLP_METRICS_PROTOCOL') ??
    getStringFromEnv('OTEL_EXPORTER_OTLP_PROTOCOL')
  )?.trim();
  if (protocol === 'grpc') {
    const { OTLPMetricExporter } =
      await import('@opentelemetry/exporter-metrics-otlp-grpc');
    return new OTLPMetricExporter();
  }
  if (protocol === 'http/json') {
    const { OTLPMetricExporter } =
      await import('@opentelemetry/exporter-metrics-otlp-http');
    return new OTLPMetricExporter();
  }
  const { OTLPMetricExporter } =
    await import('@opentelemetry/exporter-metrics-otlp-proto');
  return new OTLPMetricExporter();
}

function positiveFromEnv(name: string): number | undefined {
  const value = getNumberFromEnv(name);
  return value != null && value > 0 ? value : undefined;
}

export async function wrap() {
  if (!initialized) {
    throw new Error('Not initialized yet');
  }

  await initializeSdk();
}

export async function unwrap() {
  if (!initialized) {
    throw new Error('Not initialized yet');
  }

  if (disableInstrumentations) {
    disableInstrumentations();
    disableInstrumentations = () => {};
  }
  instrumentations = [];

  context.disable();
  propagation.disable();
  trace.disable();
  metrics.disable();
  logs.disable();
}

export async function init() {
  if (initialized) {
    return;
  }

  instrumentations = await createInstrumentations();

  // Register instrumentations synchronously to ensure code is patched even before provider is ready.
  disableInstrumentations = registerInstrumentations({
    instrumentations,
  });

  initialized = true;
}

export function logDebug(message: string, ...args: unknown[]) {
  diag.debug(message, ...args);
}

let initialized = false;
let instrumentations: Instrumentation[];
let disableInstrumentations: () => void;

// Configure lambda logging
const logLevel = diagLogLevelFromString(getStringFromEnv('OTEL_LOG_LEVEL'));
diag.setLogger(new DiagConsoleLogger(), logLevel);
