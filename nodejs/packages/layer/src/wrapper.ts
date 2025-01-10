import {
  NodeTracerConfig,
  NodeTracerProvider,
} from '@opentelemetry/sdk-trace-node';
import {
  BatchSpanProcessor,
  ConsoleSpanExporter,
  SDKRegistrationConfig,
  SimpleSpanProcessor,
} from '@opentelemetry/sdk-trace-base';
import {
  Instrumentation,
  registerInstrumentations,
} from '@opentelemetry/instrumentation';
import { awsLambdaDetector } from '@opentelemetry/resource-detector-aws';
import {
  detectResourcesSync,
  envDetector,
  processDetector,
} from '@opentelemetry/resources';
import {
  AwsInstrumentation,
  AwsSdkInstrumentationConfig,
} from '@opentelemetry/instrumentation-aws-sdk';
import {
  AwsLambdaInstrumentation,
  AwsLambdaInstrumentationConfig,
} from '@opentelemetry/instrumentation-aws-lambda';
import {
  context,
  diag,
  DiagConsoleLogger,
  DiagLogLevel,
  metrics,
  propagation,
  trace,
} from '@opentelemetry/api';
import { getEnv } from '@opentelemetry/core';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-proto';
import {
  MeterProvider,
  MeterProviderOptions,
  PeriodicExportingMetricReader,
} from '@opentelemetry/sdk-metrics';
import { OTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-proto';
import { OTLPLogExporter } from '@opentelemetry/exporter-logs-otlp-proto';
import { getPropagator } from '@opentelemetry/auto-configuration-propagators';
import {
  LoggerProvider,
  SimpleLogRecordProcessor,
  ConsoleLogRecordExporter,
  LoggerProviderConfig,
} from '@opentelemetry/sdk-logs';
import { logs } from '@opentelemetry/api-logs';

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

declare global {
  // In case of downstream configuring span processors etc
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
  function configureTracerProvider(tracerProvider: NodeTracerProvider): void;
  function configureTracer(defaultConfig: NodeTracerConfig): NodeTracerConfig;
  function configureSdkRegistration(
    defaultSdkRegistration: SDKRegistrationConfig,
  ): SDKRegistrationConfig;
  function configureInstrumentations(): Instrumentation[];
  function configureLoggerProvider(loggerProvider: LoggerProvider): void;
  function configureMeter(
    defaultConfig: MeterProviderOptions,
  ): MeterProviderOptions;
  function configureMeterProvider(meterProvider: MeterProvider): void;
  function configureLambdaInstrumentation(
    config: AwsLambdaInstrumentationConfig,
  ): AwsLambdaInstrumentationConfig;
  function configureInstrumentations(): Instrumentation[];
}

function getActiveInstumentations(): Set<string> {
  let emabledInstrumentations: string[] = defaultInstrumentationList;
  if (process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS) {
    emabledInstrumentations =
      process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS.split(',').map(i =>
        i.trim(),
      );
  }
  const instrumentationSet = new Set<string>(emabledInstrumentations);
  if (process.env.OTEL_NODE_DISABLED_INSTRUMENTATIONS) {
    const disableInstrumentations =
      process.env.OTEL_NODE_DISABLED_INSTRUMENTATIONS.split(',').map(i =>
        i.trim(),
      );
    disableInstrumentations.forEach(di => instrumentationSet.delete(di));
  }
  return instrumentationSet;
}

function defaultConfigureInstrumentations() {
  const instrumentations = [];
  const activeInstrumentations = getActiveInstumentations();
  // Use require statements for instrumentation
  // to avoid having to have transitive dependencies on all the typescript definitions.
  if (activeInstrumentations.has('dns')) {
    const {
      DnsInstrumentation,
    } = require('@opentelemetry/instrumentation-dns');
    instrumentations.push(new DnsInstrumentation());
  }
  if (activeInstrumentations.has('express')) {
    const {
      ExpressInstrumentation,
    } = require('@opentelemetry/instrumentation-express');
    instrumentations.push(new ExpressInstrumentation());
  }
  if (activeInstrumentations.has('graphql')) {
    const {
      GraphQLInstrumentation,
    } = require('@opentelemetry/instrumentation-graphql');
    instrumentations.push(new GraphQLInstrumentation());
  }
  if (activeInstrumentations.has('grpc')) {
    const {
      GrpcInstrumentation,
    } = require('@opentelemetry/instrumentation-grpc');
    instrumentations.push(new GrpcInstrumentation());
  }
  if (activeInstrumentations.has('hapi')) {
    const {
      HapiInstrumentation,
    } = require('@opentelemetry/instrumentation-hapi');
    instrumentations.push(new HapiInstrumentation());
  }
  if (activeInstrumentations.has('http')) {
    const {
      HttpInstrumentation,
    } = require('@opentelemetry/instrumentation-http');
    instrumentations.push(new HttpInstrumentation());
  }
  if (activeInstrumentations.has('ioredis')) {
    const {
      IORedisInstrumentation,
    } = require('@opentelemetry/instrumentation-ioredis');
    instrumentations.push(new IORedisInstrumentation());
  }
  if (activeInstrumentations.has('koa')) {
    const {
      KoaInstrumentation,
    } = require('@opentelemetry/instrumentation-koa');
    instrumentations.push(new KoaInstrumentation());
  }
  if (activeInstrumentations.has('mongodb')) {
    const {
      MongoDBInstrumentation,
    } = require('@opentelemetry/instrumentation-mongodb');
    instrumentations.push(new MongoDBInstrumentation());
  }
  if (activeInstrumentations.has('mysql')) {
    const {
      MySQLInstrumentation,
    } = require('@opentelemetry/instrumentation-mysql');
    instrumentations.push(new MySQLInstrumentation());
  }
  if (activeInstrumentations.has('net')) {
    const {
      NetInstrumentation,
    } = require('@opentelemetry/instrumentation-net');
    instrumentations.push(new NetInstrumentation());
  }
  if (activeInstrumentations.has('pg')) {
    const { PgInstrumentation } = require('@opentelemetry/instrumentation-pg');
    instrumentations.push(new PgInstrumentation());
  }
  if (activeInstrumentations.has('redis')) {
    const {
      RedisInstrumentation,
    } = require('@opentelemetry/instrumentation-redis');
    instrumentations.push(new RedisInstrumentation());
  }
  return instrumentations;
}

function createInstrumentations() {
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
      ? configureInstrumentations
      : defaultConfigureInstrumentations)(),
  ];
}

function initializeProvider() {
  const resource = detectResourcesSync({
    detectors: [awsLambdaDetector, envDetector, processDetector],
  });

  let config: NodeTracerConfig = {
    resource,
  };
  if (typeof configureTracer === 'function') {
    config = configureTracer(config);
  }

  const tracerProvider = new NodeTracerProvider(config);
  if (typeof configureTracerProvider === 'function') {
    configureTracerProvider(tracerProvider);
  } else {
    // Defaults
    tracerProvider.addSpanProcessor(
      new BatchSpanProcessor(new OTLPTraceExporter()),
    );
  }
  // Logging for debug
  if (logLevel === DiagLogLevel.DEBUG) {
    tracerProvider.addSpanProcessor(
      new SimpleSpanProcessor(new ConsoleSpanExporter()),
    );
  }

  let sdkRegistrationConfig: SDKRegistrationConfig = {};
  if (typeof configureSdkRegistration === 'function') {
    sdkRegistrationConfig = configureSdkRegistration(sdkRegistrationConfig);
  }
  // Auto-configure propagator if not provided
  if (!sdkRegistrationConfig.propagator) {
    sdkRegistrationConfig.propagator = getPropagator();
  }
  tracerProvider.register(sdkRegistrationConfig);

  // Configure default meter provider (doesn't export metrics)
  const metricExporter = new OTLPMetricExporter();
  let meterConfig: MeterProviderOptions = {
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

  const meterProvider = new MeterProvider(meterConfig);
  if (typeof configureMeterProvider === 'function') {
    configureMeterProvider(meterProvider);
  } else {
    metrics.setGlobalMeterProvider(meterProvider);
  }

  const logExporter = new OTLPLogExporter();
  const loggerConfig: LoggerProviderConfig = {
    resource,
  };
  const loggerProvider = new LoggerProvider(loggerConfig);
  if (typeof configureLoggerProvider === 'function') {
    configureLoggerProvider(loggerProvider);
  } else {
    loggerProvider.addLogRecordProcessor(
      new SimpleLogRecordProcessor(logExporter),
    );
    logs.setGlobalLoggerProvider(loggerProvider);
  }

  // Logging for debug
  if (logLevel === DiagLogLevel.DEBUG) {
    loggerProvider.addLogRecordProcessor(
      new SimpleLogRecordProcessor(new ConsoleLogRecordExporter()),
    );
  }

  // Create instrumentations if they have not been created before
  // to prevent additional coldstart overhead
  // caused by creations and initializations of instrumentations.
  if (!instrumentations || !instrumentations.length) {
    instrumentations = createInstrumentations();
  }

  // Re-register instrumentation with initialized provider. Patched code will see the update.
  disableInstrumentations = registerInstrumentations({
    instrumentations,
    tracerProvider,
    meterProvider,
    loggerProvider,
  });
}

export function wrap() {
  initializeProvider();
}

export function unwrap() {
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

console.log('Registering OpenTelemetry');

// Configure lambda logging
const logLevel = getEnv().OTEL_LOG_LEVEL;
diag.setLogger(new DiagConsoleLogger(), logLevel);

let instrumentations = createInstrumentations();
let disableInstrumentations: () => void;

// Register instrumentations synchronously to ensure code is patched even before provider is ready.
disableInstrumentations = registerInstrumentations({
  instrumentations,
});

wrap();
