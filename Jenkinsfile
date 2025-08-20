pipeline {
  agent any

  environment {
    environment {
  REGISTRY       = "registry.example.com"  // for Docker Hub you can use "docker.io" or leave images as USER/REPO/...
  PROJECT        = "myapp"                 // namespace/repo prefix you want
  DOCKER_CREDS   = "docker-reg-cred"       // Jenkins credentials ID
  DEPLOY_SSH_CRED = "deploy-ssh-key"       // only if using remote deploy
  DEPLOY_HOST    = "ubuntu@your-server"    // only if using remote deploy
  COMPOSE_DIR    = "/opt/myapp/deploy"     // where docker-compose.yml lives on remote
}

  }

  options { timestamps() }

  stages {
    stage('Checkout') {
      steps {
        // Use your real branch:
        git branch: 'main',
            url: 'https://github.com/paraIncog/govuemysql_tryout.git'
      }
    }

    stage('Init env') {
      steps {
        script {
          // Make sure we have a commit id and branch name
          env.GIT_SHA = sh(script: 'git rev-parse --short=7 HEAD', returnStdout: true).trim()
          env.BRANCH_NAME = env.BRANCH_NAME ?: sh(script: 'git rev-parse --abbrev-ref HEAD', returnStdout: true).trim()

          // Now compute the tag
          env.IMAGE_TAG = (env.BRANCH_NAME == 'main' || env.BRANCH_NAME == 'master' || env.BRANCH_NAME.startsWith('release/')) ?
                          "prod-${env.GIT_SHA}" : "dev-${env.GIT_SHA}"

          echo "BRANCH=${env.BRANCH_NAME}, GIT_SHA=${env.GIT_SHA}, IMAGE_TAG=${env.IMAGE_TAG}"
        }
      }
    }

    // --- Backend: test (optional) & build image ---
stage('Backend: Test') {
  steps {
    dir('backend') {
      sh 'go version || true'
      sh 'go test ./... || true'   // make strict by removing "|| true"
    }
  }
}
stage('Backend: Build Image') {
  steps {
    dir('backend') {
      sh '''
        docker build \
          --label org.opencontainers.image.revision=$GIT_SHA \
          --label org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
          -t $REGISTRY/$PROJECT/backend:$IMAGE_TAG .
      '''
    }
  }
}

// --- Frontend: build app & image ---
stage('Frontend: Build Image') {
  steps {
    dir('frontend') {
      sh '''
        npm ci
        npm run test --if-present
        npm run build --if-present
        docker build \
          --label org.opencontainers.image.revision=$GIT_SHA \
          --label org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
          -t $REGISTRY/$PROJECT/frontend:$IMAGE_TAG .
      '''
    }
  }
}

// --- Push images to registry ---
stage('Push Images') {
  steps {
    withCredentials([usernamePassword(
      credentialsId: env.DOCKER_CREDS, usernameVariable: 'USER', passwordVariable: 'PASS'
    )]) {
      sh 'echo $PASS | docker login $REGISTRY -u $USER --password-stdin'
    }
    sh '''
      docker push $REGISTRY/$PROJECT/backend:$IMAGE_TAG
      docker push $REGISTRY/$PROJECT/frontend:$IMAGE_TAG
    '''
  }
}

// --- Deploy with Docker Compose ---
// Use ONE of the options below.

stage('Deploy (Compose - same host)') {
  when { anyOf { branch 'main'; expression { env.BRANCH_NAME?.startsWith('release/') } } }
  steps {
    // Jenkins and the runtime are the same machine:
    sh '''
      export IMAGE_TAG=$IMAGE_TAG REGISTRY=$REGISTRY PROJECT=$PROJECT
      cd deploy
      docker compose pull
      docker compose up -d --remove-orphans
      docker compose ps
    '''
  }
}

// OR deploy to a remote Docker host via SSH:
stage('Deploy (Compose - remote)') {
  when { anyOf { branch 'main'; expression { env.BRANCH_NAME?.startsWith('release/') } } }
  steps {
    sshagent(credentials: [env.DEPLOY_SSH_CRED]) {
      sh """
        ssh -o StrictHostKeyChecking=no ${env.DEPLOY_HOST} \\
          'export IMAGE_TAG=${IMAGE_TAG} REGISTRY=${REGISTRY} PROJECT=${PROJECT} && \\
           cd ${COMPOSE_DIR} && docker compose pull && docker compose up -d --remove-orphans && docker compose ps'
      """
    }
  }
}

  }

  post {
    success { echo "Deployed tag ${env.IMAGE_TAG}" }
    failure { echo "Build or deploy failed." }
  }
}
