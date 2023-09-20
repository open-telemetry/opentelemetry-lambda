#!/bin/sh

cd .aws-sam/build/OTelLayer/
zip -qr ../../../<your_layer_name>.zip ruby/
cd -
