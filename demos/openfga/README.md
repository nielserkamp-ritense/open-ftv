# OpenFGA demo

This folder contains the presentation and data used for the OpenFGA demo.

## Run the OpenFGA playground

Navigate to the `open-ftv` home directory, and run the following command:
```shell
docker compose -f docker/openfga-playground.yaml up
```

Wait a little until all services run with this docker compose setup are active.

Now open your browser on http://localhost:3000/playground.

When you are finished with the playground, go back to the command-line and just press <ctrl-c> to shut down docker compose.

## Example models & tuples

To load the examples used during the presentation, run the following command while the playground is active:
```shell
cd demos/openfga
./all-examples.sh
cd -
```

## Cleanup

To clean up your docker environment:
```shell
docker compose -f docker/openfga-playground.yaml down
```

NOTE! this will delete all data.
