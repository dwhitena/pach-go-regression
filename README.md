## Pachyderm Pipline for Go-based Linear Regression

This tutorial will walk you through the implementation of training and prediction with linear regression in a Pachyderm pipeline.  Specifically, we will train a linear regression model to predict diabetes disease progression based on various medical atributes of indicators.

## The pipeline

To deploy and manage the model discussed above, we will implement it’s training, model persistence, and prediction in a Pachyderm pipeline.  This will allow us to:

- Keep a rigorous historical record of exactly what models were used on what data to produce which results.
- Automatically update online ML models when training data or parameterization changes.
- Easily revert to other versions of an ML model when a new model is not performing or when “bad data” is introduced into a training data set.

The general structure of our pipeline will look like this:

![Alt text](pipeline.jpg)

The cylinders represent data “repositories” in which Pachyderm will version training, model, etc. data (think “git for data”).  These data repositories are then input/output of the linked data processing stages (represented by the boxes in the figure).  

## Getting up and running with Pachyderm

You can experiment with this pipeline locally using a quick [local installation of Pachyderm](http://docs.pachyderm.io/en/latest/getting_started/local_installation.html).  Alternatively, you can quickly spin up a real Pachyderm cluster in any one of the popular cloud providers.  Check out the [Pachyderm docs](http://docs.pachyderm.io/en/latest/deployment/deploy_intro.html) for more details on deployment.

Once deployed, you will be able to use the Pachyderm’s `pachctl` CLI tool to create data repositories, create pipelines, and analyze our results.

## Training/fitting the linear regression model

First, let’s look our training stage.  The [goregtrain-single](goregtrain-single) and [goregtrain-multi](goregtrain-multi) go programs (and corresponding Docker images) allow us to train a single linear regression model and a multiple linear regression model, respectively, using github.com/sajari/regression.  The single linear regression model will predict diabetes disease progression based on a single attribute `bmi` (body mass index), and the multiple linear regression model will predict diabetes disease progression based on two attributes, `bmi` and `ltg` (a blood related measurement).

The data that we will use to train the model is freely available (e.g., [here](https://archive.ics.uci.edu/ml/datasets/Diabetes)) in CSV format:  

```
age,sex,bmi,map,tc,ldl,hdl,tch,ltg,glu,y
0.0380759064334,0.0506801187398,0.0616962065187,0.021872354995,-0.0442234984244,-0.0348207628377,-0.043400845652,-0.00259226199818,0.0199084208763,-0.0176461251598,151.0
-0.00188201652779,-0.044641636507,-0.0514740612388,-0.0263278347174,-0.00844872411122,-0.0191633397482,0.0744115640788,-0.0394933828741,-0.0683297436244,-0.0922040496268,75.0
0.0852989062967,0.0506801187398,0.0444512133366,-0.00567061055493,-0.0455994512826,-0.0341944659141,-0.0323559322398,-0.00259226199818,0.00286377051894,-0.0259303389895,141.0
-0.0890629393523,-0.044641636507,-0.0115950145052,-0.0366564467986,0.0121905687618,0.0249905933641,-0.0360375700439,0.0343088588777,0.0226920225667,-0.00936191133014,206.0
0.00538306037425,-0.044641636507,-0.0363846922045,0.021872354995,0.00393485161259,0.0155961395104,0.00814208360519,-0.00259226199818,-0.0319914449414,-0.0466408735636,135.0
-0.0926954778033,-0.044641636507,-0.0406959405,-0.0194420933299,-0.0689906498721,-0.0792878444118,0.041276823842,-0.07639450375,-0.041180385188,-0.0963461565417,97.0
```

The [goregtrain-single](goregtrain-single) and [goregtrain-multi](goregtrain-multi) go programs take this CSV dataset as input and output representations of the trained/fit models in a JSON format that looks like:

```
{
    "intercept": 152.13348416289583,
    "coefficients": [
        {
            "name": "bmi",
            "coefficient": 675.069774431606
        },
        {
            "name": "ltg",
            "coefficient": 614.9505047824742
        }
    ]
}
```

## Making predictions with the linear regression model.

The [goregpredict](goregpredict) go program (and corresponding Docker image) allows us to predict diabetes disease progression based on a saved JSON representation of our model (see above).  `goregpredict` takes that JSON model representation as input along with one or more JSON files, each listing particular attributes:

```
{
	"independent_variables": [
		{
			"name": "bmi",
			"value": 0.0616962065187
		},
		{
			"name": "ltg",
			"value": 0.0199084208763
		}
	]
}
```

`goregpredict` then outputs a prediction based on these attributes:

```
{
    "predicted_diabetes_progression": 210.7100380636843,
    "independent_variables": [
        {
            "name": "bmi",
            "value": 0.0616962065187
        },
        {
            "name": "ltg",
            "value": 0.0199084208763
        }
    ]
}
```

## Putting it all together, running the pipeline

First let's create Pachyderm "data repositories" in which we will version our training dataset and our attributes (from which we will make predictions):

```
➔ pachctl create-repo training
➔ pachctl create-repo attributes
➔ pachctl list-repo
NAME                CREATED             SIZE                
attributes          3 seconds ago       0 B                 
training            4 seconds ago       0 B                 
➔
```

Next we put our training data set in the `training` repo:

```
➔ cd data
➔ pachctl put-file training master -c -f diabetes.csv 
➔ pachctl list-repo
NAME                CREATED              SIZE                
training            About a minute ago   73.74 KiB           
attributes          About a minute ago   0 B                 
➔ pachctl list-file training master
NAME                TYPE                SIZE                
diabetes.csv        file                73.74 KiB           
➔
```

We can then create our training and prediction pipelines based on a [JSON specification](pipeline.json) specifying the Docker images to run for each processing stage, the input to each processing stage, and commands to run in the Docker images.  You could use either the single (`dwhitena/goregtrain:single`) or multiple (`dwhitena/goregtrain:mutli`) regression models in this specification. This will automatically trigger the training of our model and output of the JSON model representation, because Pachyderm sees that there is training data in `training` that has yet to be processed:

```
➔ cd ..
➔ pachctl create-pipeline -f pipeline.json 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                          STARTED       DURATION  RESTART PROGRESS STATE            
552ae442-6839-409e-b002-4d6fd75ad0e3 model/61e24952667842f9b0888aa69543463e 4 seconds ago 2 seconds 0       1 / 1    success 
➔ pachctl list-repo
NAME                CREATED             SIZE                
model               10 seconds ago      252 B               
prediction          10 seconds ago      0 B                 
training            3 minutes ago       73.74 KiB           
attributes          3 minutes ago       0 B                 
➔ pachctl list-file model master
NAME                TYPE                SIZE                
model.json          file                252 B               
➔ pachctl get-file model master model.json
{
    "intercept": 152.13348416289583,
    "coefficients": [
        {
            "name": "bmi",
            "coefficient": 675.069774431606
        },
        {
            "name": "ltg",
            "coefficient": 614.9505047824742
        }
    ]
}
```

Finally, we can commit some attribute files into `attributes` to trigger predictions:

```
➔ cd data/test/
➔ ls
1.json  2.json  3.json
➔ pachctl put-file attributes master -c -r -f .
➔ pachctl list-job
ID                                   OUTPUT COMMIT                               STARTED       DURATION           RESTART PROGRESS STATE            
4cf00003-af39-4488-bea0-2b9b85bffe24 prediction/655df8be34624a3190f7a85bf17d4d0d 4 seconds ago Less than a second 0       3 / 3    success 
552ae442-6839-409e-b002-4d6fd75ad0e3 model/61e24952667842f9b0888aa69543463e      3 minutes ago 2 seconds          0       1 / 1    success 
➔ pachctl list-repo
NAME                CREATED             SIZE                
prediction          3 minutes ago       803 B               
attributes          7 minutes ago       434 B               
model               3 minutes ago       252 B               
training            7 minutes ago       73.74 KiB           
➔ pachctl list-file prediction master
NAME                TYPE                SIZE                
1.json              file                267 B               
2.json              file                268 B               
3.json              file                268 B               
➔ pachctl get-file prediction master 1.json
{
    "predicted_diabetes_progression": 206.02542184806308,
    "independent_variables": [
        {
            "name": "bmi",
            "value": 0.0616962065187
        },
        {
            "name": "ltg",
            "value": 0.0199084208763
        }
    ]
}
```
