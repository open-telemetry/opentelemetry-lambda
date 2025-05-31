#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { OtelSampleLambdaStack } from '../lib/otel-sample-lambda-stack';

const app = new cdk.App();
new OtelSampleLambdaStack(app, 'OtelSampleLambdaStack');