import boto3
from threading import Thread
import os
import sys
import time

# sign of Alarm, False means no Alarm
state = False


def invoke_lambda(interval):
    while True:
        os.system("./integration.sh -n 1 -i 0")
        time.sleep(interval)
    

def alarm_puller(function_name):
    print(function_name)
    global state

    # Memory alarm
    cloudwatch.put_metric_alarm(
        AlarmName=memory_alarm,
        ActionsEnabled=True, OKActions=[], AlarmActions=[], InsufficientDataActions=[],
        MetricName='memory_utilization',
        Namespace='LambdaInsights',
        Statistic='Maximum',
        Dimensions=[
            {
                'Name': 'function_name',
                'Value': function_name
            },
        ],
        Period=60,
        EvaluationPeriods=5,
        DatapointsToAlarm=5,
        Threshold=45,
        ComparisonOperator='GreaterThanThreshold',
        TreatMissingData='notBreaching'
    )

    # CPU alarm
    cloudwatch.put_metric_alarm(
        AlarmName=cpu_alarm,
        ActionsEnabled=True, OKActions=[], AlarmActions=[], InsufficientDataActions=[],
        MetricName='cpu_total_time',
        Namespace='LambdaInsights',
        Statistic='Maximum',
        Dimensions=[
            {
                'Name': 'function_name',
                'Value': function_name
            },
        ],
        Period=60,
        EvaluationPeriods=5,
        DatapointsToAlarm=5,
        Threshold=150,
        ComparisonOperator='GreaterThanThreshold',
        TreatMissingData='notBreaching'
    )

    paginator = cloudwatch.get_paginator('describe_alarms')

    while True:
        for response in paginator.paginate(AlarmNames=[memory_alarm, cpu_alarm]):
            for alarm in response['MetricAlarms']:
                print('{} state {}'.format(alarm['AlarmName'], alarm['StateValue']))
                if alarm['StateValue'] == 'ALARM':
                    state = True
                    return
        
        time.sleep(60)

if __name__ == '__main__':

    function_name = sys.argv[1]
    print(function_name)

    soaking_time = int(sys.argv[2])
    print(soaking_time)

    emitter_interval = int(sys.argv[3])
    print(emitter_interval)

    # alarms
    memory_alarm='otel_lambda_memory-'+function_name
    cpu_alarm='otel_lambda__cpu-'+function_name

    cloudwatch = boto3.client('cloudwatch')

    # emitter thread
    Thread(target=invoke_lambda, name='emitter', args=(emitter_interval,), daemon=True).start()
    
    alarm_thread = Thread(target=alarm_puller, name='alarm_puller', args=(function_name,), daemon=True)
    alarm_thread.start()
    alarm_thread.join(soaking_time)

    # if alarm
    if state:
        print('Soaking test failed!')
        exit(1)
    else:
        print('Soaking test succeed!')
        # If no problem, delete alarm
        cloudwatch.delete_alarms(AlarmNames=[memory_alarm, cpu_alarm])
        exit(0)
