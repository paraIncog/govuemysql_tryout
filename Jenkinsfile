pipeline {
  agent any

  environment {
    REGISTRY = "registry.example.com"
    PROJECT  = "myapp"
    GIT_SHA  = "${env.GIT_COMMIT?.take(7) ?: 'local'}"
    IMAGE_TAG = "${env.BRANCH_NAME == 'main' ? 'prod-' + GIT_SHA : 'dev-' + GIT_SHA}"
  }

  options {
    skipDefaultCheckout(false)
    timestamps()
  }

  stages {
    stage('Checkout') {
      steps { checkout scm }
    }

    stage('Backend: Test & Build') {
      steps {
        dir('backend') {
          sh 'go version'
          sh 'go test ./...'
          sh 'docker build -t $REGISTRY/$PROJECT/backend:$IMAGE_TAG .'
        }
      }
    }

    stage('Frontend: Test & Build') {
      steps {
        dir('frontend') {
          sh 'node -v && npm -v'
          sh 'npm ci'
          sh 'npm run test --if-present'
          sh 'docker build -t $REGISTRY/$PROJECT/frontend:$IMAGE_TAG .'
        }
      }
    }

    stage('Login & Push Images') {
      steps {
        withCredentials([usernamePassword(credentialsId: 'docker-reg-cred', usernameVariable: 'USER', passwordVariable: 'PASS')]) {
          sh 'echo $PASS | docker login $REGISTRY -u $USER --password-stdin'
        }
        sh '''
          docker push $REGISTRY/$PROJECT/backend:$IMAGE_TAG
          docker push $REGISTRY/$PROJECT/frontend:$IMAGE_TAG
        '''
      }
    }

    stage('Deploy') {
      when { anyOf { branch 'main'; branch 'release/*' } }
      steps {
        script {
          // Option A: deploy on the SAME machine Jenkins runs on
          // sh '''
          //   export IMAGE_TAG=$IMAGE_TAG REGISTRY=$REGISTRY PROJECT=$PROJECT
          //   cd deploy && docker compose pull && docker compose up -d --remove-orphans
          // '''

          // Option B: deploy to a REMOTE Docker host via SSH
          def remoteHost = "ubuntu@your-server" // change me
          sshagent (credentials: ['deploy-ssh-key']) {
            sh """
              ssh -o StrictHostKeyChecking=no ${remoteHost} \\
                'export IMAGE_TAG=${IMAGE_TAG} REGISTRY=${REGISTRY} PROJECT=${PROJECT} && \\
                 cd /opt/myapp/deploy && docker compose pull && docker compose up -d --remove-orphans'
            """
          }
        }
      }
    }
  }

  post {
    success { echo "Deployed tag ${IMAGE_TAG}" }
    failure { echo "Build or deploy failed." }
  }
}
