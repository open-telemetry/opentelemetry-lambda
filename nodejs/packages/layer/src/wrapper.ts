import {
  context,
  diag,
  DiagConsoleLogger,
  DiagLogLevel,
  propagation,
  TextMapPropagator,
  trace,
  TracerProvider,
} from '@opentelemetry/api';
import {
  CompositePropagator,
  diagLogLevelFromString,
  getStringFromEnv,
  W3CBaggagePropagator,
  W3CTraceContextPropagator,
} from '@opentelemetry/core';
import {
  BatchSpanProcessor,
  ConsoleSpanExporter,
  NodeTracerProvider,
  SDKRegistrationConfig,
  SimpleSpanProcessor,
  SpanExporter,
  TracerConfig,
} from '@opentelemetry/sdk-trace-node';
import {
  detectResources,
  envDetector,
  Resource,
  processDetector,
} from '@opentelemetry/resources';
import { awsLambdaDetector } from '@opentelemetry/resource-detector-aws';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
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
import { AWSXRayPropagator } from '@opentelemetry/propagator-aws-xray';
import { AWSXRayLambdaPropagator } from '@opentelemetry/propagator-aws-xray-lambda';

const defaultInstrumentationList = [
  'dns',
  'express',
  'graphql',
  'grpc',
  'hapi',
  'http',
  'ioredis',
  'koa',
  'mongodb',
  'mysql',
  'net',
  'pg',
  'redis',
];

const propagatorMap = new Map<string, () => TextMapPropagator>([
  ['tracecontext', () => new W3CTraceContextPropagator()],
  ['baggage', () => new W3CBaggagePropagator()],
  ['xray', () => new AWSXRayPropagator()],
  ['xray-lambda', () => new AWSXRayLambdaPropagator()],
]);

declare global {
  // In case of downstream configuring span processors etc
  function configureLambdaInstrumentation(
    config: AwsLambdaInstrumentationConfig,
  ): AwsLambdaInstrumentationConfig;
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
  function configureInstrumentations(): Instrumentation[];
  function configureSdkRegistration(
    defaultSdkRegistration: SDKRegistrationConfig,
  ): SDKRegistrationConfig;
  function configureTracer(defaultConfig: TracerConfig): TracerConfig;

  // No explicit metric type here, but "unknown" type.
  // Because metric packages are important dynamically.
  function configureMeter(defaultConfig: unknown): unknown;
  function configureMeterProvider(meterProvider: unknown): void;

  // No explicit log type here, but "unknown" type.
  // Because log packages are important dynamically.
  function configureLoggerProvider(loggerProvider: unknown): void;
}

function getActiveInstumentations(): Set<string> {
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
  const activeInstrumentations = getActiveInstumentations();
  if (activeInstrumentations.has('amqplib')) {
    const { AmqplibInstrumentation } = await import(
      '@opentelemetry/instrumentation-amqplib'
    );
    instrumentations.push(new AmqplibInstrumentation());
  }
  if (activeInstrumentations.has('bunyan')) {
    const { BunyanInstrumentation } = await import(
      '@opentelemetry/instrumentation-bunyan'
    );
    instrumentations.push(new BunyanInstrumentation());
  }
  if (activeInstrumentations.has('cassandra-driver')) {
    const { CassandraDriverInstrumentation } = await import(
      '@opentelemetry/instrumentation-cassandra-driver'
    );
    instrumentations.push(new CassandraDriverInstrumentation());
  }
  if (activeInstrumentations.has('connect')) {
    const { ConnectInstrumentation } = await import(
      '@opentelemetry/instrumentation-connect'
    );
    instrumentations.push(new ConnectInstrumentation());
  }
  if (activeInstrumentations.has('dataloader')) {
    const { DataloaderInstrumentation } = await import(
      '@opentelemetry/instrumentation-dataloader'
    );
    instrumentations.push(new DataloaderInstrumentation());
  }
  if (activeInstrumentations.has('dns')) {
    const { DnsInstrumentation } = await import(
      '@opentelemetry/instrumentation-dns'
    );
    instrumentations.push(new DnsInstrumentation());
  }
  if (activeInstrumentations.has('express')) {
    const { ExpressInstrumentation } = await import(
      '@opentelemetry/instrumentation-express'
    );
    instrumentations.push(new ExpressInstrumentation());
  }
  if (activeInstrumentations.has('fs')) {
    const { FsInstrumentation } = await import(
      '@opentelemetry/instrumentation-fs'
    );
    instrumentations.push(new FsInstrumentation());
  }
  if (activeInstrumentations.has('graphql')) {
    const { GraphQLInstrumentation } = await import(
      '@opentelemetry/instrumentation-graphql'
    );
    instrumentations.push(new GraphQLInstrumentation());
  }
  if (activeInstrumentations.has('grpc')) {
    const { GrpcInstrumentation } = await import(
      '@opentelemetry/instrumentation-grpc'
    );
    instrumentations.push(new GrpcInstrumentation());
  }
  if (activeInstrumentations.has('hapi')) {
    const { HapiInstrumentation } = await import(
      '@opentelemetry/instrumentation-hapi'
    );
    instrumentations.push(new HapiInstrumentation());
  }
  if (activeInstrumentations.has('http')) {
    const { HttpInstrumentation } = await import(
      '@opentelemetry/instrumentation-http'
    );
    instrumentations.push(new HttpInstrumentation());
  }
  if (activeInstrumentations.has('ioredis')) {
    const { IORedisInstrumentation } = await import(
      '@opentelemetry/instrumentation-ioredis'
    );
    instrumentations.push(new IORedisInstrumentation());
  }
  if (activeInstrumentations.has('kafkajs')) {
    const { KafkaJsInstrumentation } = await import(
      '@opentelemetry/instrumentation-kafkajs'
    );
    instrumentations.push(new KafkaJsInstrumentation());
  }
  if (activeInstrumentations.has('knex')) {
    const { KnexInstrumentation } = await import(
      '@opentelemetry/instrumentation-knex'
    );
    instrumentations.push(new KnexInstrumentation());
  }
  if (activeInstrumentations.has('koa')) {
    const { KoaInstrumentation } = await import(
      '@opentelemetry/instrumentation-koa'
    );
    instrumentations.push(new KoaInstrumentation());
  }
  if (activeInstrumentations.has('memcached')) {
    const { MemcachedInstrumentation } = await import(
      '@opentelemetry/instrumentation-memcached'
    );
    instrumentations.push(new MemcachedInstrumentation());
  }
  if (activeInstrumentations.has('mongodb')) {
    const { MongoDBInstrumentation } = await import(
      '@opentelemetry/instrumentation-mongodb'
    );
    instrumentations.push(new MongoDBInstrumentation());
  }
  if (activeInstrumentations.has('mongoose')) {
    const { MongooseInstrumentation } = await import(
      '@opentelemetry/instrumentation-mongoose'
    );
    instrumentations.push(new MongooseInstrumentation());
  }
  if (activeInstrumentations.has('mysql')) {
    const { MySQLInstrumentation } = await import(
      '@opentelemetry/instrumentation-mysql'
    );
    instrumentations.push(new MySQLInstrumentation());
  }
  if (activeInstrumentations.has('mysql2')) {
    const { MySQL2Instrumentation } = await import(
      '@opentelemetry/instrumentation-mysql2'
    );
    instrumentations.push(new MySQL2Instrumentation());
  }
  if (activeInstrumentations.has('nestjs-core')) {
    const { NestInstrumentation } = await import(
      '@opentelemetry/instrumentation-nestjs-core'
    );
    instrumentations.push(new NestInstrumentation());
  }
  if (activeInstrumentations.has('net')) {
    const { NetInstrumentation } = await import(
      '@opentelemetry/instrumentation-net'
    );
    instrumentations.push(new NetInstrumentation());
  }
  if (activeInstrumentations.has('pg')) {
    const { PgInstrumentation } = await import(
      '@opentelemetry/instrumentation-pg'
    );
    instrumentations.push(new PgInstrumentation());
  }
  if (activeInstrumentations.has('pino')) {
    const { PinoInstrumentation } = await import(
      '@opentelemetry/instrumentation-pino'
    );
    instrumentations.push(new PinoInstrumentation());
  }
  if (activeInstrumentations.has('redis')) {
    const { RedisInstrumentation } = await import(
      '@opentelemetry/instrumentation-redis'
    );
    instrumentations.push(new RedisInstrumentation());
  }
  if (activeInstrumentations.has('restify')) {
    const { RestifyInstrumentation } = await import(
      '@opentelemetry/instrumentation-restify'
    );
    instrumentations.push(new RestifyInstrumentation());
  }
  if (activeInstrumentations.has('socket.io')) {
    const { SocketIoInstrumentation } = await import(
      '@opentelemetry/instrumentation-socket.io'
    );
    instrumentations.push(new SocketIoInstrumentation());
  }
  if (activeInstrumentations.has('undici')) {
    const { UndiciInstrumentation } = await import(
      '@opentelemetry/instrumentation-undici'
    );
    instrumentations.push(new UndiciInstrumentation());
  }
  if (activeInstrumentations.has('winston')) {
    const { WinstonInstrumentation } = await import(
      '@opentelemetry/instrumentation-winston'
    );
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

function getPropagator(): TextMapPropagator {
  if (
    process.env.OTEL_PROPAGATORS == null ||
    process.env.OTEL_PROPAGATORS.trim() === ''
  ) {
    return new CompositePropagator({
      propagators: [
        new W3CTraceContextPropagator(),
        new W3CBaggagePropagator(),
      ],
    });
  }
  const propagatorsFromEnv = Array.from(
    new Set(
      process.env.OTEL_PROPAGATORS?.split(',').map(value =>
        value.toLowerCase().trim(),
      ),
    ),
  );
  const propagators = propagatorsFromEnv.flatMap(propagatorName => {
    if (propagatorName === 'none') {
      diag.info(
        'Not selecting any propagator for value "none" specified in the environment variable OTEL_PROPAGATORS',
      );
      return [];
    }
    const propagatorFactoryFunction = propagatorMap.get(propagatorName);
    if (propagatorFactoryFunction == null) {
      diag.warn(
        `Invalid propagator "${propagatorName}" specified in the environment variable OTEL_PROPAGATORS`,
      );
      return [];
    }
    return propagatorFactoryFunction();
  });
  return new CompositePropagator({ propagators });
}

function getExportersFromEnv(): SpanExporter[] | null {
  if (
    process.env.OTEL_TRACES_EXPORTER == null ||
    process.env.OTEL_TRACES_EXPORTER.trim() === ''
  ) {
    return [];
  }
  if (process.env.OTEL_TRACES_EXPORTER.includes('none')) {
    return null;
  }

  const stringToExporter = new Map<string, () => SpanExporter>([
    ['otlp', () => new OTLPTraceExporter()],
    ['console', () => new ConsoleSpanExporter()],
  ]);
  const exporters: SpanExporter[] = [];
  process.env.OTEL_TRACES_EXPORTER.split(',').map(exporterName => {
    exporterName = exporterName.toLowerCase().trim();
    const exporter = stringToExporter.get(exporterName);
    if (exporter) {
      exporters.push(exporter());
    } else {
      diag.warn(
        `Invalid exporter "${exporterName}" specified in the environment variable OTEL_TRACES_EXPORTER`,
      );
    }
  });
  return exporters;
}

async function initializeTracerProvider(
  resource: Resource,
): Promise<TracerProvider | undefined> {
  let config: TracerConfig = {
    resource,
    spanProcessors: [],
  };

  const exporters = getExportersFromEnv();
  if (!exporters) {
    return;
  }

  if (typeof configureTracer === 'function') {
    config = configureTracer(config);
  }

  if (exporters.length) {
    config.spanProcessors = [];
    exporters.map(exporter => {
      if (exporter instanceof ConsoleSpanExporter) {
        config.spanProcessors?.push(new SimpleSpanProcessor(exporter));
      } else {
        config.spanProcessors?.push(new BatchSpanProcessor(exporter));
      }
    });
  }

  config.spanProcessors = config.spanProcessors || [];
  if (config.spanProcessors.length === 0) {
    // Default
    config.spanProcessors.push(new BatchSpanProcessor(new OTLPTraceExporter()));
  }

  // Logging for debug
  if (logLevel === DiagLogLevel.DEBUG) {
    config.spanProcessors.push(
      new SimpleSpanProcessor(new ConsoleSpanExporter()),
    );
  }

  const tracerProvider = new NodeTracerProvider(config);

  let sdkRegistrationConfig: SDKRegistrationConfig = {};
  if (typeof configureSdkRegistration === 'function') {
    sdkRegistrationConfig = configureSdkRegistration(sdkRegistrationConfig);
  }
  // Auto-configure propagator if not provided
  if (!sdkRegistrationConfig.propagator) {
    sdkRegistrationConfig.propagator = getPropagator();
  }
  tracerProvider.register(sdkRegistrationConfig);

  return tracerProvider;
}

async function initializeMeterProvider(
  resource: Resource,
): Promise<unknown | undefined> {
  if (process.env.OTEL_METRICS_EXPORTER === 'none') {
    return;
  }

  const { metrics } = await import('@opentelemetry/api');
  const { MeterProvider, PeriodicExportingMetricReader } = await import(
    '@opentelemetry/sdk-metrics'
  );
  const { OTLPMetricExporter } = await import(
    '@opentelemetry/exporter-metrics-otlp-http'
  );

  // Configure default meter provider (doesn't export metrics)
  const metricExporter = new OTLPMetricExporter();
  let meterConfig: unknown = {
    resource,
    readers: [
      new PeriodicExportingMetricReader({
        exporter: metricExporter,
      }),
    ],
  };
  if (typeof configureMeter === 'function') {
    meterConfig = configureMeter(meterConfig);
  }

  const meterProvider = new MeterProvider(meterConfig as object);
  if (typeof configureMeterProvider === 'function') {
    configureMeterProvider(meterProvider);
  } else {
    metrics.setGlobalMeterProvider(meterProvider);
  }

  metricsDisableFunction = () => {
    metrics.disable();
  };

  return meterProvider;
}

async function initializeLoggerProvider(
  resource: Resource,
): Promise<unknown | undefined> {
  if (process.env.OTEL_LOGS_EXPORTER === 'none') {
    return;
  }

  const { logs } = await import('@opentelemetry/api-logs');
  const {
    LoggerProvider,
    BatchLogRecordProcessor,
    SimpleLogRecordProcessor,
    ConsoleLogRecordExporter,
  } = await import('@opentelemetry/sdk-logs');
  const { OTLPLogExporter } = await import(
    '@opentelemetry/exporter-logs-otlp-http'
  );

  const logExporter = new OTLPLogExporter();
  const loggerConfig = {
    resource,
  };
  const loggerProvider = new LoggerProvider(loggerConfig);
  if (typeof configureLoggerProvider === 'function') {
    configureLoggerProvider(loggerProvider);
  } else {
    loggerProvider.addLogRecordProcessor(
      new BatchLogRecordProcessor(logExporter),
    );
    logs.setGlobalLoggerProvider(loggerProvider);
  }

  // Logging for debug
  if (logLevel === DiagLogLevel.DEBUG) {
    loggerProvider.addLogRecordProcessor(
      new SimpleLogRecordProcessor(new ConsoleLogRecordExporter()),
    );
  }

  logsDisableFunction = () => {
    logs.disable();
  };

  return loggerProvider;
}

async function initializeProvider() {
  const resource = detectResources({
    detectors: [awsLambdaDetector, envDetector, processDetector],
  });

  const tracerProvider: TracerProvider | undefined =
    await initializeTracerProvider(resource);
  const meterProvider: unknown | undefined =
    await initializeMeterProvider(resource);
  const loggerProvider: unknown | undefined =
    await initializeLoggerProvider(resource);

  // Create instrumentations if they have not been created before
  // to prevent additional coldstart overhead
  // caused by creations and initializations of instrumentations.
  if (!instrumentations || !instrumentations.length) {
    instrumentations = await createInstrumentations();
  }

  // Re-register instrumentation with initialized provider. Patched code will see the update.

  disableInstrumentations = registerInstrumentations({
    instrumentations,
    tracerProvider,
    // eslint-disable-next-line  @typescript-eslint/no-explicit-any
    meterProvider: meterProvider as any,
    // eslint-disable-next-line  @typescript-eslint/no-explicit-any
    loggerProvider: loggerProvider as any,
  });
}

export async function wrap() {
  if (!initialized) {
    throw new Error('Not initialized yet');
  }

  await initializeProvider();
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

  if (metricsDisableFunction) {
    metricsDisableFunction();
    metricsDisableFunction = () => {};
  }

  if (logsDisableFunction) {
    logsDisableFunction();
    logsDisableFunction = () => {};
  }
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

console.log('Registering OpenTelemetry');

let initialized = false;
let instrumentations: Instrumentation[];
let disableInstrumentations: () => void;
let metricsDisableFunction: () => void;
let logsDisableFunction: () => void;

// Configure lambda logging
const logLevel = diagLogLevelFromString(getStringFromEnv('OTEL_LOG_LEVEL'));
diag.setLogger(new DiagConsoleLogger(), logLevel);
