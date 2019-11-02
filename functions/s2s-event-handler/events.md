Example event message received by event handler lambda:

```
map[
    Records:[
        map[
            EventSource:aws:sns 
            EventSubscriptionArn:arn:aws:sns:eu-west-1:014341863605:s2s-stack-events:826e4271-5104-4962-af87-d1b78d3f4957 
            EventVersion:1.0 
            Sns:map[
                Message:
                    StackId='arn:aws:cloudformation:eu-west-1:014341863605:stack/s2s/20d4f1e0-f8e0-11e9-9b71-0a0cb138cf0a'
                    Timestamp='2019-10-27T17:36:11.613Z'
                    EventId='cgw0b574d54e07b6445d-CREATE_IN_PROGRESS-2019-10-27T17:36:11.613Z'
                    LogicalResourceId='cgw0b574d54e07b6445d'
                    Namespace='014341863605'
                    ResourceProperties='
                    {
                        "Type": "ipsec.1",
                        "IpAddress": "94.226.70.22",
                        "Tags": [
                            {
                                "Value": "ducbase-home",
                                "Key": "Name"
                            },
                            {
                                "Value": "site-to-site",
                                "Key": "Application"
                            }
                        ],
                        "BgpAsn": "65011"
                    }'
                    ResourceStatus='CREATE_IN_PROGRESS'
                    ResourceStatusReason=''
                    ResourceType='AWS::EC2::CustomerGateway'
                    StackName='s2s'
                    ClientRequestToken='null'
                MessageAttributes:map[] 
                MessageId:5b8d72b7-0e25-5d47-8dab-2498d9ee69e2 
                Signature:Sx7LfVzGOJ0VksQnoj++gy8lW675lHsfdqQBHyPOMeuC18DTP1N0Q3JegpHQOcUHVXdD4m+plwkCS5lquD+XGD/Qqy+cCCi78YRlwfwxsQbrYQKSGn+5LgiEwdUrwURhJKcBMPLCQaT9L+P948Vys3RDquKd4NXR1yCDkjDy0YOeh+MmTGLsSxYIfgMLdqaEgSZTCVhjsE+7GKwn81Bq3OR4GcpnCc9qy213YbHmQvQGdBmfOOrX3PZzbXi4X1DMysm7ra5AzYcgYoVxEE0Q9f3cD74ZLajCJF9L9xpTT37vKcvkybgViHBzcmKNjvrhuLe2FrQ/buxo6S6/bfm+Jw== 
                SignatureVersion:1 
                SigningCertUrl:https://sns.eu-west-1.amazonaws.com/SimpleNotificationService-6aad65c2f9911b05cd53efda11f913f9.pem 
                Subject:AWS CloudFormation Notification 
                Timestamp:2019-10-27T17:36:11.697Z 
                TopicArn:arn:aws:sns:eu-west-1:014341863605:s2s-stack-events 
                Type:Notification 
                UnsubscribeUrl:https://sns.eu-west-1.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:eu-west-1:014341863605:s2s-stack-events:826e4271-5104-4962-af87-d1b78d3f4957
            ]
        ]
    ]
]
```